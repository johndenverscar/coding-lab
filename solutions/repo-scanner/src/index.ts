import { Command } from 'commander';
import { SecretScanner } from './scanner';
import { Reporter } from './reporter';
import { GitHubClient } from './github-client';
import { ScanOptions } from './types';
import * as dotenv from 'dotenv';

dotenv.config();

interface ScanCliOptions extends ScanOptions {
    token?: string;
    concurrency?: string;
}

interface ScanResult {
    exitCode: number;
    error?: Error;
}

/**
 * Performs a repository scan and generates a report.
 * Returns an exit code (0 for success, 1 for secrets found or error).
 * This function can be used independently of the CLI.
 */
export async function performScan(options: ScanCliOptions): Promise<ScanResult> {
    try {
        const concurrency = options.concurrency ? parseInt(options.concurrency, 10) : 10;
        const githubClient = new GitHubClient(options.token);
        const scanner = new SecretScanner(githubClient, [], concurrency);

        console.log(`Starting scan of ${options.owner}/${options.repo}/${options.branch || 'main'}...`);

        const results = await scanner.scanRepository({
            owner: options.owner,
            repo: options.repo,
            branch: options.branch
        });

        Reporter.generateReport(results);

        // Return exit code based on findings
        return {
            exitCode: results.length > 0 ? 1 : 0
        };
    } catch (error) {
        console.error('Error:', error);
        return {
            exitCode: 1,
            error: error instanceof Error ? error : new Error(String(error))
        };
    }
}

const program = new Command();

program
    .name('github-secret-scanner')
    .description('Scan GitHub repositories for exposed secrets')
    .version('1.0.0');

program
    .command('scan')
    .description('Scan a GitHub repository')
    .requiredOption('-o, --owner <owner>', 'Repository owner')
    .requiredOption('-r, --repo <repo>', 'Repository name')
    .option('-b, --branch <branch>', 'Branch to scan', 'main')
    .option('-t, --token <token>', 'GitHub token (or use GITHUB_TOKEN env)')
    .option('-c, --concurrency <number>', 'Number of concurrent file requests', '10')
    .action(async (options: ScanCliOptions) => {
        const result = await performScan(options);
        process.exit(result.exitCode);
    });

program.parse(process.argv);