module.exports = {
    globDirectory: 'dist',
    globPatterns: [
        '**/*.{js,css,html,ico,json}',
    ],
    swDest: 'dist/service-worker.js',
    sourcemap: false,
    cleanupOutdatedCaches: true,
    skipWaiting: true,
    clientsClaim: true,
};
