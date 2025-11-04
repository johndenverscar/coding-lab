import { ScanResult, Severity } from './types';

export interface IReporter {
    generateReport(results: ScanResult[]): void;
}

export class ConsoleReporter implements IReporter {
    generateReport(results: ScanResult[]): void {
        if (results.length === 0) {
            console.log('\nNo secrets found!');
            return;
        }

        console.log(`\nFound ${results.length} potential secret(s):\n`);

        const grouped = this.groupBySeverity(results);

        for (const severity of [Severity.HIGH, Severity.MEDIUM, Severity.LOW]) {
            const items = grouped[severity];
            if (items.length === 0) continue;

            console.log(`${severity.toUpperCase()} SEVERITY (${items.length}):`);

            for (const result of items) {
                console.log(`\n  File: ${result.file}:${result.line}`);
                console.log(`  Type: ${result.type}`);
                console.log(`  Match: ${result.match}`);
                console.log(`  Context:\n${this.indent(result.context, 4)}`);
            }
            console.log('');
        }
    }

    private groupBySeverity(results: ScanResult[]) {
        return {
            [Severity.HIGH]: results.filter(r => r.severity === Severity.HIGH),
            [Severity.MEDIUM]: results.filter(r => r.severity === Severity.MEDIUM),
            [Severity.LOW]: results.filter(r => r.severity === Severity.LOW)
        };
    }

    private indent(text: string, spaces: number): string {
        const padding = ' '.repeat(spaces);
        return text.split('\n').map(line => padding + line).join('\n');
    }
}

export class JSONReporter implements IReporter {
    generateReport(results: ScanResult[]): void {
        const report = {
            summary: {
                totalFindings: results.length,
                highSeverity: results.filter(r => r.severity === Severity.HIGH).length,
                mediumSeverity: results.filter(r => r.severity === Severity.MEDIUM).length,
                lowSeverity: results.filter(r => r.severity === Severity.LOW).length
            },
            findings: results.map(r => ({
                file: r.file,
                line: r.line,
                type: r.type,
                severity: r.severity,
                match: r.match,
                context: r.context
            }))
        };

        console.log(JSON.stringify(report, null, 2));
    }
}

export class SARIFReporter implements IReporter {
    generateReport(results: ScanResult[]): void {
        const sarif = {
            version: "2.1.0",
            $schema: "https://raw.githubusercontent.com/oasis-tcs/sarif-spec/master/Schemata/sarif-schema-2.1.0.json",
            runs: [{
                tool: {
                    driver: {
                        name: "github-secret-scanner",
                        version: "1.0.0",
                        informationUri: "https://github.com/example/repo-scanner"
                    }
                },
                results: results.map(r => ({
                    ruleId: this.getRuleId(r.type),
                    level: this.mapSeverity(r.severity),
                    message: {
                        text: `Potential ${r.type} detected`
                    },
                    locations: [{
                        physicalLocation: {
                            artifactLocation: {
                                uri: r.file
                            },
                            region: {
                                startLine: r.line,
                                snippet: {
                                    text: r.match
                                }
                            }
                        }
                    }]
                }))
            }]
        };

        console.log(JSON.stringify(sarif, null, 2));
    }

    private getRuleId(type: string): string {
        return type.toLowerCase().replace(/\s+/g, '-');
    }

    private mapSeverity(severity: Severity): string {
        const severityMap = {
            [Severity.HIGH]: 'error',
            [Severity.MEDIUM]: 'warning',
            [Severity.LOW]: 'note'
        };
        return severityMap[severity];
    }
}

// Legacy static class for backward compatibility
export class Reporter {
    static generateReport(results: ScanResult[]): void {
        const reporter = new ConsoleReporter();
        reporter.generateReport(results);
    }
}