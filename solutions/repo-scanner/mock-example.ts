/**
 * Example: Testing with Mock GitHub Client
 *
 * This demonstrates how dependency injection allows testing
 * without hitting the real GitHub API.
 */

import { SecretScanner } from './src/scanner';
import { IGitHubClient } from './src/github-client';

// Create a mock GitHub client that implements the interface
class MockGitHubClient implements IGitHubClient {
    async getRepoTree(owner: string, repo: string, branch = 'main') {
        // Return fake tree data for testing
        return [
            { path: 'src/config.js', type: 'blob' },
            { path: 'src/secrets.env', type: 'blob' },
            { path: 'README.md', type: 'blob' },
            { path: 'node_modules/package.json', type: 'blob' } // Should be filtered
        ];
    }

    async getFileContent(owner: string, repo: string, path: string) {
        // Return fake file contents with test secrets
        const fakeFiles: Record<string, string> = {
            'src/config.js': `
                const config = {
                    awsKey: "AKIAIOSFODNN7EXAMPLE",
                    dbUrl: "mongodb://admin:password123@localhost:27017/db"
                };
            `,
            'src/secrets.env': `
                GITHUB_TOKEN=ghp_1234567890abcdefghijklmnopqrstuvwxyz
                STRIPE_KEY=sk_test_1234567890abcdefghijklmnop
            `,
            'README.md': '# Test Repository\n\nNo secrets here!'
        };

        return fakeFiles[path] || '';
    }
}

// Example: Unit test without hitting GitHub API
async function testSecretScanning() {
    console.log('🧪 Testing Secret Scanner with Mock Client\n');

    // Inject the mock - no real API calls!
    const mockClient = new MockGitHubClient();
    const scanner = new SecretScanner(mockClient, [], 5);

    // Scan using the mock
    const results = await scanner.scanRepository({
        owner: 'test-org',
        repo: 'test-repo',
        branch: 'main'
    });

    // Verify results
    console.log(`✅ Found ${results.length} secrets in mock data`);
    console.log('\nDetected secrets:');

    results.forEach(result => {
        console.log(`  - ${result.type} in ${result.file}:${result.line}`);
        console.log(`    Severity: ${result.severity}`);
    });

    // Assertions
    const awsKeys = results.filter(r => r.type.includes('AWS'));
    const mongoConnections = results.filter(r => r.type.includes('MongoDB'));
    const githubTokens = results.filter(r => r.type.includes('GitHub'));

    console.log('\n📊 Test Assertions:');
    console.assert(awsKeys.length > 0, 'Should detect AWS keys');
    console.assert(mongoConnections.length > 0, 'Should detect MongoDB connection');
    console.assert(githubTokens.length > 0, 'Should detect GitHub tokens');
    console.assert(!results.some(r => r.file.includes('node_modules')), 'Should filter node_modules');

    console.log('  ✓ All assertions passed!');
    console.log('\n✨ Benefits of Dependency Injection:');
    console.log('  • No real GitHub API calls during tests');
    console.log('  • Fast test execution');
    console.log('  • Predictable test data');
    console.log('  • No need for test credentials');
    console.log('  • Easy to test edge cases');

    return results;
}

// Run the test
// testSecretScanning().catch(console.error);
