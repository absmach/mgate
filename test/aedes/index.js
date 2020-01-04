var aedes = require('aedes')()
var server = require('net').createServer(aedes.handle)
var port = 1884

server.listen(port, function () {
  console.log('server listening on port', port)
})

aedes.on('subscribe', function (subscriptions, client) {
    console.log('SUBSCRIBE: client \x1b[32m' + (client ? client.id : client) +
        '\x1b[0m subscribed to topics: ' + subscriptions.map(s => s.topic).join('\n'), 'from broker', aedes.id)
})

aedes.on('unsubscribe', function (subscriptions, client) {
    console.log('UNSUBSCRIBE: client \x1b[32m' + (client ? client.id : client) +
        '\x1b[0m unsubscribed to topics: ' + subscriptions.join('\n'), 'from broker', aedes.id)
})

// fired when a client connects
aedes.on('client', function (client) {
    console.log('Client Connected: \x1b[33m' + (client ? client.id : client) + '\x1b[0m', 'to broker', aedes.id)
})

// fired when a client disconnects
aedes.on('clientDisconnect', function (client) {
    console.log('Client Disconnected: \x1b[31m' + (client ? client.id : client) + '\x1b[0m', 'to broker', aedes.id)
})

// fired when a message is published
aedes.on('publish', async function (packet, client) {
    console.log('PUBLISH: client \x1b[31m' + (client ? client.id : 'BROKER_' + aedes.id) + '\x1b[0m has published', packet.payload.toString(), 'on', packet.topic, 'to broker', aedes.id)
})

aedes.on('clientError', function (client, err) {
    console.log('client error: client: %s, error: %s', client.id, err.message);
});

aedes.on('connectionError', function (client, err) {
    console.log('connection error: client: %s, error: %s', client.id, err.message);
});

aedes.on('error', function (err) {
    console.log('aedes error: %s', err.message);
});