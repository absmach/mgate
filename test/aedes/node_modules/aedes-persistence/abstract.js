'use strict'

var concat = require('concat-stream')
var through = require('through2')
var Packet = require('aedes-packet')

function abstractPersistence (opts) {
  var test = opts.test
  var _persistence = opts.persistence
  var waitForReady = opts.waitForReady

  // requiring it here so it will not error for modules
  // not using the default emitter
  var buildEmitter = opts.buildEmitter || require('mqemitter')

  if (_persistence.length === 0) {
    _persistence = function asyncify (cb) {
      cb(null, opts.persistence())
    }
  }

  function persistence (cb) {
    var mq = buildEmitter()
    var broker = {
      id: 'broker-42',
      mq: mq,
      publish: mq.emit.bind(mq),
      subscribe: mq.on.bind(mq),
      unsubscribe: mq.removeListener.bind(mq)
    }

    _persistence(function (err, instance) {
      if (instance) {
        // Wait for ready event, if applicable, to ensure the persistence isn't
        // destroyed while it's still being set up.
        // https://github.com/mcollina/aedes-persistence-redis/issues/41
        if (waitForReady) {
          // We have to listen to 'ready' before setting broker because that
          // can result in 'ready' being emitted.
          instance.on('ready', function () {
            instance.removeListener('error', cb)
            cb(null, instance)
          })
          instance.on('error', cb)
        }
        instance.broker = broker
        if (waitForReady) {
          // 'ready' event will call back.
          return
        }
      }
      cb(err, instance)
    })
  }

  function storeRetained (instance, opts, cb) {
    opts = opts || {}

    var packet = {
      cmd: 'publish',
      id: instance.broker.id,
      topic: opts.topic || 'hello/world',
      payload: opts.payload || Buffer.from('muahah'),
      qos: 0,
      retain: true
    }

    instance.storeRetained(packet, function (err) {
      cb(err, packet)
    })
  }

  function matchRetainedWithPattern (t, pattern, opts) {
    persistence(function (err, instance) {
      if (err) { throw err }

      storeRetained(instance, opts, function (err, packet) {
        t.notOk(err, 'no error')
        var stream
        if (Array.isArray(pattern)) {
          stream = instance.createRetainedStreamCombi(pattern)
        } else {
          stream = instance.createRetainedStream(pattern)
        }

        stream.pipe(concat(function (list) {
          t.deepEqual(list, [packet], 'must return the packet')
          instance.destroy(t.end.bind(t))
        }))
      })
    })
  }

  function testInstance (title, cb) {
    test(title, function (t) {
      persistence(function (err, instance) {
        if (err) { throw err }
        cb(t, instance)
      })
    })
  }

  test('store and look up retained messages', function (t) {
    matchRetainedWithPattern(t, 'hello/world')
  })

  test('look up retained messages with a # pattern', function (t) {
    matchRetainedWithPattern(t, '#')
  })

  test('look up retained messages with a + pattern', function (t) {
    matchRetainedWithPattern(t, 'hello/+')
  })

  test('look up retained messages with multiple patterns', function (t) {
    matchRetainedWithPattern(t, ['hello/+', 'other/hello'])
  })

  testInstance('remove retained message', function (t, instance) {
    storeRetained(instance, {}, function (err, packet) {
      t.notOk(err, 'no error')
      storeRetained(instance, {
        payload: Buffer.alloc(0)
      }, function (err) {
        t.notOk(err, 'no error')

        var stream = instance.createRetainedStream('#')

        stream.pipe(concat(function (list) {
          t.deepEqual(list, [], 'must return an empty list')
          instance.destroy(t.end.bind(t))
        }))
      })
    })
  })

  testInstance('storing twice a retained message should keep only the last', function (t, instance) {
    storeRetained(instance, {}, function (err, packet) {
      t.notOk(err, 'no error')
      storeRetained(instance, {
        payload: Buffer.from('ahah')
      }, function (err, packet) {
        t.notOk(err, 'no error')

        var stream = instance.createRetainedStream('#')

        stream.pipe(concat(function (list) {
          t.deepEqual(list, [packet], 'must return the last packet')
          instance.destroy(t.end.bind(t))
        }))
      })
    })
  })

  testInstance('Create a new packet while storing a retained message', function (t, instance) {
    var packet = {
      cmd: 'publish',
      id: instance.broker.id,
      topic: opts.topic || 'hello/world',
      payload: opts.payload || Buffer.from('muahah'),
      qos: 0,
      retain: true
    }
    var newPacket = Object.assign({}, packet)

    instance.storeRetained(packet, function (err) {
      t.notOk(err, 'no error')
      // packet reference change to check if a new packet is stored always
      packet.retain = false
      var stream = instance.createRetainedStream('#')

      stream.pipe(concat(function (list) {
        t.deepEqual(list, [newPacket], 'must return the last packet')
        instance.destroy(t.end.bind(t))
      }))
    })
  })

  testInstance('store and look up subscriptions by client', function (t, instance) {
    var client = { id: 'abcde' }
    var subs = [{
      topic: 'hello',
      qos: 1
    }, {
      topic: 'matteo',
      qos: 1
    }, {
      topic: 'noqos',
      qos: 0
    }]

    instance.addSubscriptions(client, subs, function (err, reClient) {
      t.equal(reClient, client, 'client must be the same')
      t.notOk(err, 'no error')
      instance.subscriptionsByClient(client, function (err, resubs, reReClient) {
        t.equal(reReClient, client, 'client must be the same')
        t.notOk(err, 'no error')
        t.deepEqual(resubs, subs)
        instance.destroy(t.end.bind(t))
      })
    })
  })

  testInstance('remove subscriptions by client', function (t, instance) {
    var client = { id: 'abcde' }
    var subs = [{
      topic: 'hello',
      qos: 1
    }, {
      topic: 'matteo',
      qos: 1
    }]

    instance.addSubscriptions(client, subs, function (err, reClient) {
      t.notOk(err, 'no error')
      instance.removeSubscriptions(client, ['hello'], function (err, reClient) {
        t.notOk(err, 'no error')
        t.equal(reClient, client, 'client must be the same')
        instance.subscriptionsByClient(client, function (err, resubs, reClient) {
          t.equal(reClient, client, 'client must be the same')
          t.notOk(err, 'no error')
          t.deepEqual(resubs, [{
            topic: 'matteo',
            qos: 1
          }])
          instance.destroy(t.end.bind(t))
        })
      })
    })
  })

  testInstance('store and look up subscriptions by topic', function (t, instance) {
    var client = { id: 'abcde' }
    var subs = [{
      topic: 'hello',
      qos: 1
    }, {
      topic: 'hello/#',
      qos: 1
    }, {
      topic: 'matteo',
      qos: 1
    }]

    instance.addSubscriptions(client, subs, function (err) {
      t.notOk(err, 'no error')
      instance.subscriptionsByTopic('hello', function (err, resubs) {
        t.notOk(err, 'no error')
        t.deepEqual(resubs, [{
          clientId: client.id,
          topic: 'hello/#',
          qos: 1
        }, {
          clientId: client.id,
          topic: 'hello',
          qos: 1
        }])
        instance.destroy(t.end.bind(t))
      })
    })
  })

  testInstance('get client list after subscriptions', function (t, instance) {
    var client1 = { id: 'abcde' }
    var client2 = { id: 'efghi' }
    var subs = [{
      topic: 'helloagain',
      qos: 1
    }]

    instance.addSubscriptions(client1, subs, function (err) {
      t.notOk(err, 'no error for client 1')
      instance.addSubscriptions(client2, subs, function (err) {
        t.notOk(err, 'no error for client 2')
        var stream = instance.getClientList(subs[0].topic)
        stream.pipe(concat({ encoding: 'object' }, function (out) {
          t.deepEqual(out, [client1.id, client2.id])
          instance.destroy(t.end.bind(t))
        }))
      })
    })
  })

  testInstance('get client list after an unsubscribe', function (t, instance) {
    var client1 = { id: 'abcde' }
    var client2 = { id: 'efghi' }
    var subs = [{
      topic: 'helloagain',
      qos: 1
    }]

    instance.addSubscriptions(client1, subs, function (err) {
      t.notOk(err, 'no error for client 1')
      instance.addSubscriptions(client2, subs, function (err) {
        t.notOk(err, 'no error for client 2')
        instance.removeSubscriptions(client2, [subs[0].topic], function (err, reClient) {
          t.notOk(err, 'no error for removeSubscriptions')
          var stream = instance.getClientList(subs[0].topic)
          stream.pipe(concat({ encoding: 'object' }, function (out) {
            t.deepEqual(out, [client1.id])
            instance.destroy(t.end.bind(t))
          }))
        })
      })
    })
  })

  testInstance('get subscriptions list after an unsubscribe', function (t, instance) {
    var client1 = { id: 'abcde' }
    var client2 = { id: 'efghi' }
    var subs = [{
      topic: 'helloagain',
      qos: 1
    }]

    instance.addSubscriptions(client1, subs, function (err) {
      t.notOk(err, 'no error for client 1')
      instance.addSubscriptions(client2, subs, function (err) {
        t.notOk(err, 'no error for client 2')
        instance.removeSubscriptions(client2, [subs[0].topic], function (err, reClient) {
          t.notOk(err, 'no error for removeSubscriptions')
          instance.subscriptionsByTopic(subs[0].topic, function (err, clients) {
            t.notOk(err, 'no error getting subscriptions by topic')
            t.deepEqual(clients[0].clientId, client1.id)
            instance.destroy(t.end.bind(t))
          })
        })
      })
    })
  })

  testInstance('QoS 0 subscriptions, restored but not matched', function (t, instance) {
    var client = { id: 'abcde' }
    var subs = [{
      topic: 'hello',
      qos: 0
    }, {
      topic: 'hello/#',
      qos: 1
    }, {
      topic: 'matteo',
      qos: 1
    }]

    instance.addSubscriptions(client, subs, function (err) {
      t.notOk(err, 'no error')
      instance.subscriptionsByClient(client, function (err, resubs) {
        t.notOk(err, 'no error')
        t.deepEqual(resubs, subs)
        instance.subscriptionsByTopic('hello', function (err, resubs2) {
          t.notOk(err, 'no error')
          t.deepEqual(resubs2, [{
            clientId: client.id,
            topic: 'hello/#',
            qos: 1
          }])
          instance.destroy(t.end.bind(t))
        })
      })
    })
  })

  testInstance('clean subscriptions', function (t, instance) {
    var client = { id: 'abcde' }
    var subs = [{
      topic: 'hello',
      qos: 1
    }, {
      topic: 'matteo',
      qos: 1
    }]

    instance.addSubscriptions(client, subs, function (err) {
      t.notOk(err, 'no error')
      instance.cleanSubscriptions(client, function (err) {
        t.notOk(err, 'no error')
        instance.subscriptionsByTopic('hello', function (err, resubs) {
          t.notOk(err, 'no error')
          t.deepEqual(resubs, [], 'no subscriptions')

          instance.subscriptionsByClient(client, function (err, resubs) {
            t.error(err)
            t.deepEqual(resubs, null, 'no subscriptions')

            instance.countOffline(function (err, subsCount, clientsCount) {
              t.error(err, 'no error')
              t.equal(subsCount, 0, 'no subscriptions added')
              t.equal(clientsCount, 0, 'no clients added')

              instance.destroy(t.end.bind(t))
            })
          })
        })
      })
    })
  })

  testInstance('clean subscriptions with no active subscriptions', function (t, instance) {
    var client = { id: 'abcde' }

    instance.cleanSubscriptions(client, function (err) {
      t.notOk(err, 'no error')
      instance.subscriptionsByTopic('hello', function (err, resubs) {
        t.notOk(err, 'no error')
        t.deepEqual(resubs, [], 'no subscriptions')

        instance.subscriptionsByClient(client, function (err, resubs) {
          t.error(err)
          t.deepEqual(resubs, null, 'no subscriptions')

          instance.countOffline(function (err, subsCount, clientsCount) {
            t.error(err, 'no error')
            t.equal(subsCount, 0, 'no subscriptions added')
            t.equal(clientsCount, 0, 'no clients added')

            instance.destroy(t.end.bind(t))
          })
        })
      })
    })
  })

  testInstance('same topic, different QoS', function (t, instance) {
    var client = { id: 'abcde' }
    var subs = [{
      topic: 'hello',
      qos: 0
    }, {
      topic: 'hello',
      qos: 1
    }]

    instance.addSubscriptions(client, subs, function (err, reClient) {
      t.equal(reClient, client, 'client must be the same')
      t.error(err, 'no error')

      instance.subscriptionsByClient(client, function (err, subsForClient, client) {
        t.error(err, 'no error')
        t.deepEqual(subsForClient, [{
          topic: 'hello',
          qos: 1
        }])

        instance.subscriptionsByTopic('hello', function (err, subsForTopic) {
          t.error(err, 'no error')
          t.deepEqual(subsForTopic, [{
            clientId: 'abcde',
            topic: 'hello',
            qos: 1
          }])

          instance.countOffline(function (err, subsCount, clientsCount) {
            t.error(err, 'no error')
            t.equal(subsCount, 1, 'one subscription added')
            t.equal(clientsCount, 1, 'one client added')

            instance.destroy(t.end.bind(t))
          })
        })
      })
    })
  })

  testInstance('replace subscriptions', function (t, instance) {
    var client = { id: 'abcde' }
    var topic = 'hello'
    var sub = { topic: topic }
    var subByTopic = { clientId: client.id, topic: topic }

    function check (qos, cb) {
      sub.qos = subByTopic.qos = qos
      instance.addSubscriptions(client, [sub], function (err, reClient) {
        t.equal(reClient, client, 'client must be the same')
        t.error(err, 'no error')
        instance.subscriptionsByClient(client, function (err, subsForClient, client) {
          t.error(err, 'no error')
          t.deepEqual(subsForClient, [sub])
          instance.subscriptionsByTopic(topic, function (err, subsForTopic) {
            t.error(err, 'no error')
            t.deepEqual(subsForTopic, qos === 0 ? [] : [subByTopic])
            instance.countOffline(function (err, subsCount, clientsCount) {
              t.error(err, 'no error')
              if (qos === 0) {
                t.equal(subsCount, 0, 'no subscriptions added')
              } else {
                t.equal(subsCount, 1, 'one subscription added')
              }
              t.equal(clientsCount, 1, 'one client added')
              cb()
            })
          })
        })
      })
    }

    check(0, function () {
      check(1, function () {
        check(2, function () {
          check(1, function () {
            check(0, function () {
              instance.destroy(t.end.bind(t))
            })
          })
        })
      })
    })
  })

  testInstance('replace subscriptions in same call', function (t, instance) {
    var client = { id: 'abcde' }
    var topic = 'hello'
    var subs = [
      { topic: topic, qos: 0 },
      { topic: topic, qos: 1 },
      { topic: topic, qos: 2 },
      { topic: topic, qos: 1 },
      { topic: topic, qos: 0 }
    ]
    instance.addSubscriptions(client, subs, function (err, reClient) {
      t.equal(reClient, client, 'client must be the same')
      t.error(err, 'no error')
      instance.subscriptionsByClient(client, function (err, subsForClient, client) {
        t.error(err, 'no error')
        t.deepEqual(subsForClient, [{ topic: topic, qos: 0 }])
        instance.subscriptionsByTopic(topic, function (err, subsForTopic) {
          t.error(err, 'no error')
          t.deepEqual(subsForTopic, [])
          instance.countOffline(function (err, subsCount, clientsCount) {
            t.error(err, 'no error')
            t.equal(subsCount, 0, 'no subscriptions added')
            t.equal(clientsCount, 1, 'one client added')
            instance.destroy(t.end.bind(t))
          })
        })
      })
    })
  })

  testInstance('store and count subscriptions', function (t, instance) {
    var client = { id: 'abcde' }
    var subs = [{
      topic: 'hello',
      qos: 1
    }, {
      topic: 'matteo',
      qos: 1
    }, {
      topic: 'noqos',
      qos: 0
    }]

    instance.addSubscriptions(client, subs, function (err, reClient) {
      t.equal(reClient, client, 'client must be the same')
      t.error(err, 'no error')

      instance.countOffline(function (err, subsCount, clientsCount) {
        t.error(err, 'no error')
        t.equal(subsCount, 2, 'two subscriptions added')
        t.equal(clientsCount, 1, 'one client added')

        instance.removeSubscriptions(client, ['hello'], function (err, reClient) {
          t.error(err, 'no error')

          instance.countOffline(function (err, subsCount, clientsCount) {
            t.error(err, 'no error')
            t.equal(subsCount, 1, 'one subscription added')
            t.equal(clientsCount, 1, 'one client added')

            instance.removeSubscriptions(client, ['matteo'], function (err, reClient) {
              t.error(err, 'no error')

              instance.countOffline(function (err, subsCount, clientsCount) {
                t.error(err, 'no error')
                t.equal(subsCount, 0, 'zero subscriptions added')
                t.equal(clientsCount, 1, 'one client added')

                instance.removeSubscriptions(client, ['noqos'], function (err, reClient) {
                  t.error(err, 'no error')

                  instance.countOffline(function (err, subsCount, clientsCount) {
                    t.error(err, 'no error')
                    t.equal(subsCount, 0, 'zero subscriptions added')
                    t.equal(clientsCount, 0, 'zero clients added')

                    instance.removeSubscriptions(client, ['noqos'], function (err, reClient) {
                      t.error(err, 'no error')

                      instance.countOffline(function (err, subsCount, clientsCount) {
                        t.error(err, 'no error')
                        t.equal(subsCount, 0, 'zero subscriptions added')
                        t.equal(clientsCount, 0, 'zero clients added')

                        instance.destroy(t.end.bind(t))
                      })
                    })
                  })
                })
              })
            })
          })
        })
      })
    })
  })

  testInstance('count subscriptions with two clients', function (t, instance) {
    var client1 = { id: 'abcde' }
    var client2 = { id: 'fghij' }
    var subs = [{
      topic: 'hello',
      qos: 1
    }, {
      topic: 'matteo',
      qos: 1
    }, {
      topic: 'noqos',
      qos: 0
    }]

    function remove (client, subs, expectedSubs, expectedClients, cb) {
      instance.removeSubscriptions(client, subs, function (err, reClient) {
        t.error(err, 'no error')
        t.equal(reClient, client, 'client must be the same')

        instance.countOffline(function (err, subsCount, clientsCount) {
          t.error(err, 'no error')
          t.equal(subsCount, expectedSubs, 'subscriptions added')
          t.equal(clientsCount, expectedClients, 'clients added')

          cb()
        })
      })
    }

    instance.addSubscriptions(client1, subs, function (err, reClient) {
      t.equal(reClient, client1, 'client must be the same')
      t.error(err, 'no error')

      instance.addSubscriptions(client2, subs, function (err, reClient) {
        t.equal(reClient, client2, 'client must be the same')
        t.error(err, 'no error')

        remove(client1, ['foobar'], 4, 2, function () {
          remove(client1, ['hello'], 3, 2, function () {
            remove(client1, ['hello'], 3, 2, function () {
              remove(client1, ['matteo'], 2, 2, function () {
                remove(client1, ['noqos'], 2, 1, function () {
                  remove(client2, ['hello'], 1, 1, function () {
                    remove(client2, ['matteo'], 0, 1, function () {
                      remove(client2, ['noqos'], 0, 0, function () {
                        instance.destroy(t.end.bind(t))
                      })
                    })
                  })
                })
              })
            })
          })
        })
      })
    })
  })

  testInstance('add duplicate subs to persistence for qos > 0', function (t, instance) {
    var client = { id: 'abcde' }
    var topic = 'hello'
    var subs = [{
      topic: topic,
      qos: 1
    }]

    instance.addSubscriptions(client, subs, function (err, reClient) {
      t.equal(reClient, client, 'client must be the same')
      t.error(err, 'no error')

      instance.addSubscriptions(client, subs, function (err, resCLient) {
        t.equal(resCLient, client, 'client must be the same')
        t.error(err, 'no error')
        subs[0].clientId = client.id
        instance.subscriptionsByTopic(topic, function (err, subsForTopic) {
          t.error(err, 'no error')
          t.deepEqual(subsForTopic, subs)
          instance.destroy(t.end.bind(t))
        })
      })
    })
  })

  testInstance('add duplicate subs to persistence for qos 0', function (t, instance) {
    var client = { id: 'abcde' }
    var topic = 'hello'
    var subs = [{
      topic: topic,
      qos: 0
    }]

    instance.addSubscriptions(client, subs, function (err, reClient) {
      t.equal(reClient, client, 'client must be the same')
      t.error(err, 'no error')

      instance.addSubscriptions(client, subs, function (err, resCLient) {
        t.equal(resCLient, client, 'client must be the same')
        t.error(err, 'no error')
        instance.subscriptionsByClient(client, function (err, subsForClient, client) {
          t.error(err, 'no error')
          t.deepEqual(subsForClient, subs)
          instance.destroy(t.end.bind(t))
        })
      })
    })
  })

  testInstance('get topic list after concurrent subscriptions of a client', function (t, instance) {
    var client = { id: 'abcde' }
    var subs1 = [{
      topic: 'hello1',
      qos: 1
    }]
    var subs2 = [{
      topic: 'hello2',
      qos: 1
    }]
    var calls = 2

    function done () {
      if (!--calls) {
        instance.subscriptionsByClient(client, function (err, resubs) {
          t.notOk(err, 'no error')
          t.deepEqual(resubs, [subs1[0], subs2[0]])
          instance.destroy(t.end.bind(t))
        })
      }
    }

    instance.addSubscriptions(client, subs1, function (err) {
      t.notOk(err, 'no error for hello1')
      done()
    })
    instance.addSubscriptions(client, subs2, function (err) {
      t.notOk(err, 'no error for hello2')
      done()
    })
  })

  testInstance('add outgoing packet and stream it', function (t, instance) {
    var sub = {
      clientId: 'abcde',
      topic: 'hello',
      qos: 1
    }
    var client = {
      id: sub.clientId
    }
    var packet = {
      cmd: 'publish',
      topic: 'hello',
      payload: Buffer.from('world'),
      qos: 1,
      dup: false,
      length: 14,
      retain: false,
      brokerId: instance.broker.id,
      brokerCounter: 42
    }
    var expected = {
      cmd: 'publish',
      topic: 'hello',
      payload: Buffer.from('world'),
      qos: 1,
      retain: false,
      brokerId: instance.broker.id,
      brokerCounter: 42,
      messageId: 0
    }

    instance.outgoingEnqueue(sub, packet, function (err) {
      t.error(err)
      var stream = instance.outgoingStream(client)

      stream.pipe(concat(function (list) {
        t.deepEqual(list, [expected], 'must return the packet')
        instance.destroy(t.end.bind(t))
      }))
    })
  })

  testInstance('add outgoing packet for multiple subs and stream to all', function (t, instance) {
    var sub = {
      clientId: 'abcde',
      topic: 'hello',
      qos: 1
    }
    var sub2 = {
      clientId: 'fghih',
      topic: 'hello',
      qos: 1
    }
    var subs = [sub, sub2]
    var client = {
      id: sub.clientId
    }
    var client2 = {
      id: sub2.clientId
    }
    var packet = {
      cmd: 'publish',
      topic: 'hello',
      payload: Buffer.from('world'),
      qos: 1,
      dup: false,
      length: 14,
      retain: false,
      brokerId: instance.broker.id,
      brokerCounter: 42
    }
    var expected = {
      cmd: 'publish',
      topic: 'hello',
      payload: Buffer.from('world'),
      qos: 1,
      retain: false,
      brokerId: instance.broker.id,
      brokerCounter: 42,
      messageId: 0
    }

    instance.outgoingEnqueueCombi(subs, packet, function (err) {
      t.error(err)
      var stream = instance.outgoingStream(client)
      stream.pipe(concat(function (list) {
        t.deepEqual(list, [expected], 'must return the packet')

        var stream2 = instance.outgoingStream(client2)
        stream2.pipe(concat(function (list) {
          t.deepEqual(list, [expected], 'must return the packet')
          instance.destroy(t.end.bind(t))
        }))
      }))
    })
  })

  testInstance('add outgoing packet as a string and stream', function (t, instance) {
    var sub = {
      clientId: 'abcde',
      topic: 'hello',
      qos: 1
    }
    var client = {
      id: sub.clientId
    }
    var packet = {
      cmd: 'publish',
      topic: 'hello',
      payload: 'world',
      qos: 1,
      dup: false,
      length: 14,
      retain: false,
      brokerId: instance.broker.id,
      brokerCounter: 42
    }
    var expected = {
      cmd: 'publish',
      topic: 'hello',
      payload: 'world',
      qos: 1,
      retain: false,
      brokerId: instance.broker.id,
      brokerCounter: 42,
      messageId: 0
    }

    instance.outgoingEnqueueCombi([sub], packet, function (err) {
      t.error(err)
      var stream = instance.outgoingStream(client)

      stream.pipe(concat(function (list) {
        t.deepEqual(list, [expected], 'must return the packet')
        instance.destroy(t.end.bind(t))
      }))
    })
  })

  testInstance('add outgoing packet and stream it twice', function (t, instance) {
    var sub = {
      clientId: 'abcde',
      topic: 'hello',
      qos: 1
    }
    var client = {
      id: sub.clientId
    }
    var packet = {
      cmd: 'publish',
      topic: 'hello',
      payload: Buffer.from('world'),
      qos: 1,
      dup: false,
      length: 14,
      retain: false,
      brokerId: instance.broker.id,
      brokerCounter: 42,
      messageId: 4242
    }
    var expected = {
      cmd: 'publish',
      topic: 'hello',
      payload: Buffer.from('world'),
      qos: 1,
      retain: false,
      brokerId: instance.broker.id,
      brokerCounter: 42,
      messageId: 0
    }

    instance.outgoingEnqueueCombi([sub], packet, function (err) {
      t.error(err)
      var stream = instance.outgoingStream(client)

      stream.pipe(concat(function (list) {
        t.deepEqual(list, [expected], 'must return the packet')

        var stream = instance.outgoingStream(client)

        stream.pipe(concat(function (list) {
          t.deepEqual(list, [expected], 'must return the packet')
          t.notEqual(list[0], expected, 'packet must be a different object')
          instance.destroy(t.end.bind(t))
        }))
      }))
    })
  })

  function enqueueAndUpdate (t, instance, client, sub, packet, messageId, callback) {
    instance.outgoingEnqueueCombi([sub], packet, function (err) {
      t.error(err)
      var updated = new Packet(packet)
      updated.messageId = messageId

      instance.outgoingUpdate(client, updated, function (err, reclient, repacket) {
        t.error(err)
        t.equal(reclient, client, 'client matches')
        t.equal(repacket, repacket, 'packet matches')
        callback(updated)
      })
    })
  }

  testInstance('add outgoing packet and update messageId', function (t, instance) {
    var sub = {
      clientId: 'abcde', topic: 'hello', qos: 1
    }
    var client = {
      id: sub.clientId
    }
    var packet = {
      cmd: 'publish',
      topic: 'hello',
      payload: Buffer.from('world'),
      qos: 1,
      dup: false,
      length: 14,
      retain: false,
      brokerId: instance.broker.id,
      brokerCounter: 42
    }

    enqueueAndUpdate(t, instance, client, sub, packet, 42, function (updated) {
      var stream = instance.outgoingStream(client)
      delete updated.messageId
      stream.pipe(concat(function (list) {
        delete list[0].messageId
        t.notEqual(list[0], updated, 'must not be the same object')
        t.deepEqual(list, [updated], 'must return the packet')
        instance.destroy(t.end.bind(t))
      }))
    })
  })

  testInstance('add 2 outgoing packet and clear messageId', function (t, instance) {
    var sub = {
      clientId: 'abcde', topic: 'hello', qos: 1
    }
    var client = {
      id: sub.clientId
    }
    var packet1 = {
      cmd: 'publish',
      topic: 'hello',
      payload: Buffer.from('world'),
      qos: 1,
      dup: false,
      length: 14,
      retain: false,
      brokerId: instance.broker.id,
      brokerCounter: 42
    }
    var packet2 = {
      cmd: 'publish',
      topic: 'hello',
      payload: Buffer.from('matteo'),
      qos: 1,
      dup: false,
      length: 14,
      retain: false,
      brokerId: instance.broker.id,
      brokerCounter: 43
    }

    enqueueAndUpdate(t, instance, client, sub, packet1, 42, function (updated1) {
      enqueueAndUpdate(t, instance, client, sub, packet2, 43, function (updated2) {
        instance.outgoingClearMessageId(client, updated1, function (err, packet) {
          t.error(err)
          t.deepEqual(packet.messageId, 42, 'must have the same messageId')
          t.deepEqual(packet.payload.toString(), packet1.payload.toString(), 'must have original payload')
          t.deepEqual(packet.topic, packet1.topic, 'must have original topic')
          var stream = instance.outgoingStream(client)
          delete updated2.messageId
          stream.pipe(concat(function (list) {
            delete list[0].messageId
            t.notEqual(list[0], updated2, 'must not be the same object')
            t.deepEqual(list, [updated2], 'must return the packet')
            instance.destroy(t.end.bind(t))
          }))
        })
      })
    })
  })

  testInstance('update to pubrel', function (t, instance) {
    var sub = {
      clientId: 'abcde', topic: 'hello', qos: 1
    }
    var client = {
      id: sub.clientId
    }
    var packet = {
      cmd: 'publish',
      topic: 'hello',
      payload: Buffer.from('world'),
      qos: 2,
      dup: false,
      length: 14,
      retain: false,
      brokerId: instance.broker.id,
      brokerCounter: 42
    }

    instance.outgoingEnqueueCombi([sub], packet, function (err) {
      t.error(err)
      var updated = new Packet(packet)
      updated.messageId = 42

      instance.outgoingUpdate(client, updated, function (err, reclient, repacket) {
        t.error(err)
        t.equal(reclient, client, 'client matches')
        t.equal(repacket, repacket, 'packet matches')

        var pubrel = {
          cmd: 'pubrel',
          messageId: updated.messageId
        }

        instance.outgoingUpdate(client, pubrel, function (err) {
          t.error(err)

          var stream = instance.outgoingStream(client)

          stream.pipe(concat(function (list) {
            t.deepEqual(list, [pubrel], 'must return the packet')
            instance.destroy(t.end.bind(t))
          }))
        })
      })
    })
  })

  testInstance('add incoming packet, get it, and clear with messageId', function (t, instance) {
    var client = {
      id: 'abcde'
    }
    var packet = {
      cmd: 'publish',
      topic: 'hello',
      payload: Buffer.from('world'),
      qos: 2,
      dup: false,
      length: 14,
      retain: false,
      messageId: 42
    }

    instance.incomingStorePacket(client, packet, function (err) {
      t.error(err)

      instance.incomingGetPacket(client, {
        messageId: packet.messageId
      }, function (err, retrieved) {
        t.error(err)

        // adjusting the objects so they match
        delete retrieved.brokerCounter
        delete retrieved.brokerId
        delete packet.dup
        delete packet.length

        t.deepEqual(retrieved, packet, 'retrieved packet must be deeply equal')
        t.notEqual(retrieved, packet, 'retrieved packet must not be the same objet')

        instance.incomingDelPacket(client, retrieved, function (err) {
          t.error(err)

          instance.incomingGetPacket(client, {
            messageId: packet.messageId
          }, function (err, retrieved) {
            t.ok(err, 'must error')
            instance.destroy(t.end.bind(t))
          })
        })
      })
    })
  })

  testInstance('store, fetch and delete will message', function (t, instance) {
    var client = {
      id: '12345'
    }
    var expected = {
      topic: 'hello/died',
      payload: Buffer.from('muahahha'),
      qos: 0,
      retain: true
    }

    instance.putWill(client, expected, function (err, c) {
      t.error(err, 'no error')
      t.equal(c, client, 'client matches')
      instance.getWill(client, function (err, packet, c) {
        t.error(err, 'no error')
        t.deepEqual(packet, expected, 'will matches')
        t.equal(c, client, 'client matches')
        client.brokerId = packet.brokerId
        instance.delWill(client, function (err, packet, c) {
          t.error(err, 'no error')
          t.deepEqual(packet, expected, 'will matches')
          t.equal(c, client, 'client matches')
          instance.getWill(client, function (err, packet, c) {
            t.error(err, 'no error')
            t.notOk(packet, 'no will after del')
            t.equal(c, client, 'client matches')
            instance.destroy(t.end.bind(t))
          })
        })
      })
    })
  })

  testInstance('stream all will messages', function (t, instance) {
    var client = {
      id: '12345'
    }
    var toWrite = {
      topic: 'hello/died',
      payload: Buffer.from('muahahha'),
      qos: 0,
      retain: true
    }

    instance.putWill(client, toWrite, function (err, c) {
      t.error(err, 'no error')
      t.equal(c, client, 'client matches')
      instance.streamWill().pipe(through.obj(function (chunk, enc, cb) {
        t.deepEqual(chunk, {
          clientId: client.id,
          brokerId: instance.broker.id,
          topic: 'hello/died',
          payload: Buffer.from('muahahha'),
          qos: 0,
          retain: true
        }, 'packet matches')
        cb()
        client.brokerId = chunk.brokerId
        instance.delWill(client, function (err, result, client) {
          t.error(err, 'no error')
          instance.destroy(t.end.bind(t))
        })
      }))
    })
  })

  testInstance('stream all will message for unknown brokers', function (t, instance) {
    var originalId = instance.broker.id
    var client = {
      id: '42'
    }
    var anotherClient = {
      id: '24'
    }
    var toWrite1 = {
      topic: 'hello/died42',
      payload: Buffer.from('muahahha'),
      qos: 0,
      retain: true
    }
    var toWrite2 = {
      topic: 'hello/died24',
      payload: Buffer.from('muahahha'),
      qos: 0,
      retain: true
    }

    instance.putWill(client, toWrite1, function (err, c) {
      t.error(err, 'no error')
      t.equal(c, client, 'client matches')
      instance.broker.id = 'anotherBroker'
      instance.putWill(anotherClient, toWrite2, function (err, c) {
        t.error(err, 'no error')
        t.equal(c, anotherClient, 'client matches')
        instance.streamWill({
          'anotherBroker': Date.now()
        })
          .pipe(through.obj(function (chunk, enc, cb) {
            t.deepEqual(chunk, {
              clientId: client.id,
              brokerId: originalId,
              topic: 'hello/died42',
              payload: Buffer.from('muahahha'),
              qos: 0,
              retain: true
            }, 'packet matches')
            cb()
            client.brokerId = chunk.brokerId
            instance.delWill(client, function (err, result, client) {
              t.error(err, 'no error')
              instance.destroy(t.end.bind(t))
            })
          }))
      })
    })
  })

  testInstance('delete wills from dead brokers', function (t, instance) {
    var client = {
      id: '42'
    }

    var toWrite1 = {
      topic: 'hello/died42',
      payload: Buffer.from('muahahha'),
      qos: 0,
      retain: true
    }

    instance.putWill(client, toWrite1, function (err, c) {
      t.error(err, 'no error')
      t.equal(c, client, 'client matches')
      instance.broker.id = 'anotherBroker'
      client.brokerId = instance.broker.id
      instance.delWill(client, function (err, result, client) {
        t.error(err, 'no error')
        instance.destroy(t.end.bind(t))
      })
    })
  })

  testInstance('do not error if unkown messageId in outoingClearMessageId', function (t, instance) {
    var client = {
      id: 'abc-123'
    }

    instance.outgoingClearMessageId(client, 42, function (err) {
      t.error(err)
      instance.destroy(t.end.bind(t))
    })
  })
}

module.exports = abstractPersistence
