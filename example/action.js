// Action Script
// Test: node action.js --name=test
// Send action: --action="Test|chanify://action/run-script/test?name=test"

const chanify = require('chanify');  // Load chanify module
console.log('show name', chanify.args['name']); // Test get argument
chanify.routeTo('shortcuts://'); // Test jump to shortcut app
