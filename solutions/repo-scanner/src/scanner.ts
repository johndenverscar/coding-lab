import { GitHubClient, IGitHubClient } from "./github-client";
import { SECRET_PATTERNS } from "./patterns";
import { ScanResult, ScanOptions } from "./types";

export class SecretScanner {
    private githubClient: IGitHubClient;
    private excludePatterns: RegExp[];
    private concurrencyLimit: number;

    constructor(
        githubClient: IGitHubClient,
        excludePatterns: RegExp[] = [],
        concurrencyLimit: number = 10
    ) {
        this.githubClient = githubClient;
        this.excludePatterns = excludePatterns;
        this.concurrencyLimit = concurrencyLimit;
    }

    setExcludePatterns(patterns: RegExp[]) {
        this.excludePatterns = patterns;
    }

    private chunkArray<T>(array: T[], size: number): T[][] {
        const chunks: T[][] = [];
        for (let i = 0; i < array.length; i += size) {
            chunks.push(array.slice(i, i + size));
        }
        return chunks;
    }

    private async scanFileWithRetry(owner: string, repo: string, filePath: string): Promise<ScanResult[]> {
        try {
            const content = await this.githubClient.getFileContent(owner, repo, filePath);
            return this.scanContent(content, filePath);
        } catch (error) {
            console.warn(`Failed to fetch or scan file ${filePath}: ${(error as Error).message}`);
            return [];
        }
    }

    private shouldExclude(filePath: string): boolean {
        // Always exclude common non-scannable files
        const defaultExcludes = [
            /node_modules\//,
            /\.git\//,
            /dist\//,
            /build\//,
            /\.min\.js$/,
            /\.map$/,
            /package-lock\.json$/,
            /yarn\.lock$/
        ];

        return [...defaultExcludes, ...this.excludePatterns]
            .some(pattern => pattern.test(filePath));
    }

    async scanRepository(options: ScanOptions): Promise<ScanResult[]> {
        const results: ScanResult[] = [];
        const { owner, repo, branch = 'main' } = options;

        console.log(`Fetching repository tree for ${owner}/${repo}...`);
        const tree = await this.githubClient.getRepoTree(owner, repo, branch);

        const files = tree.filter((item: { path?: string, type?: string }) =>
            item.type === 'blob' && item.path && !this.shouldExclude(item.path)
        );

        console.log(`Scanning ${files.length} files for secrets (${this.concurrencyLimit} concurrent requests)...`);

        // Process files in parallel batches
        const filePaths = files
            .map((f: { path?: string }) => f.path)
            .filter((path: string | undefined): path is string => !!path);
        const chunks: string[][] = this.chunkArray(filePaths, this.concurrencyLimit);

        for (const chunk of chunks) {
            const chunkResults = await Promise.all(
                chunk.map(filePath => this.scanFileWithRetry(owner, repo, filePath))
            );
            results.push(...chunkResults.flat());
        }

        return results;
    }

    private scanContent(content: string, filePath: string): ScanResult[] {
        const results: ScanResult[] = [];
        const lines = content.split('\n');

        for (const pattern of SECRET_PATTERNS) {
            for (let i = 0; i < lines.length; i++) {
                const line = lines[i];
                const matches = line.matchAll(pattern.pattern);
                for (const match of matches) {
                    results.push({
                        file: filePath,
                        line: i + 1,
                        match: match[0],
                        type: pattern.name,
                        severity: pattern.severity,
                        context: this.getContext(lines, i)
                    });
                }
            }
        }

        return results;
    }

    private getContext(lines: string[], lineIndex: number): string {
        const start = Math.max(0, lineIndex - 1);
        const end = Math.min(lines.length, lineIndex + 2);
        return lines.slice(start, end).join('\n');
    }
}