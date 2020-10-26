/**
 * Mock .svg imports
 */

const path = require('path')

module.exports = {
  process(_, filename) {
    return 'module.exports = "' + path.basename(filename) + '.svg"'
  },
  getCacheKey(_, filename) {
    // The output is based on path.
    return path.basename(filename)
  },
}
