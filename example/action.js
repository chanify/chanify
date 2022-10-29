// Action Script
// Test: node action.js --name=test
// Send action: --action="Test|chanify://action/run-script/test?name=test"

// Load chanify module
const chanify = require('chanify');

// Test write & read pasteboard
chanify.pasteboard = 123;
chanify.alert({
    title: 'Title',
    message: chanify.pasteboard,
    action: 'ok'
}, () => {
    console.log("ok");
});

// Test get argument
//console.log('show name', chanify.args['name']);
/*
// Test http get
const http = require('https');

const postData = JSON.stringify({
    'text': 'Hello World!'
});

const options = {
    hostname: 'api.chanify.net',
    path: '/v1/sender/CID07JYGEiJBQzNETlRWQklVR1FMUUY0UjQ3UFlCVElGTUtVU0RUQTdNIgIIAQ.j1Pn5gPMNX-GUDjyCnY7052niFZcrHB-w-THRSW9fxI',
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'Content-Length': Buffer.byteLength(postData)
    }
};
const req = http.request(options, res => {
    let data = '';
    res.on('data', chunk => {
        data += chunk;
    });
    res.on('end', () => {
        console.log(`STATUS: ${res.statusCode}`);
        //console.log('res:', data);
    });
});

req.on('error', (e) => {
    console.error(`problem with request: ${e.message}`);
});
  
// Write data to request body
req.write(postData);
req.end();

*/

// http.get('http://www.baidu.com', res => {
//     let data = '';
//     res.on('data', chunk => {
//         data += chunk;
//     });
//     res.on('end', () => {
//         console.log('Response ended:', data.toString());
//     });
// }).on('error', err => {
//     console.error(err);
// });

// Test jump to shortcut app
//chanify.routeTo('shortcuts://');
