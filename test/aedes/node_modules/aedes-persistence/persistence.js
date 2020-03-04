'use strict'

var from2 = require('from2')
var QlobberSub = require('qlobber/aedes/qlobber-sub')
var QlobberTrue = require('qlobber').QlobberTrue
var Packet = require('aedes-packet')
var QlobberOpts = {
  wildcard_one: '+',
  wildcard_some: '#',
  separator: '/'
}

function MemoryPersistence () {
  if (!(this instanceof MemoryPersistence)) {
    return new MemoryPersistence()
  }

  this._retained = []
  // clientId -> topic -> qos
  this._subscriptions = new Map()
  this._clientsCount = 0
  this._trie = new QlobberSub(QlobberOpts)
  this._outgoing = {}
  this._incoming = {}
  this._wills = {}
}

function matchTopic (p) {
  return p.topic !== this.topic
}

MemoryPersistence.prototype.storeRetained = function (packet, cb) {
  packet = Object.assign({}, packet)
  this._retained = this._retained.filter(matchTopic, packet)

  if (packet.payload.length > 0) this._retained.push(packet)

  cb(null)
}

function matchingStream (current, pattern) {
  var matcher = new QlobberTrue(QlobberOpts)

  if (Array.isArray(pattern)) {
    for (var i = 0; i < pattern.length; i += 1) {
      matcher.add(pattern[i])
    }
  } else {
    matcher.add(pattern)
  }

  return from2.obj(function match (size, next) {
    var entry

    while ((entry = current.shift()) != null) {
      if (matcher.test(entry.topic)) {
        setImmediate(next, null, entry)
        return
      }
    }

    if (!entry) this.push(null)
  })
}

MemoryPersistence.prototype.createRetainedStream = function (pattern) {
  return matchingStream([].concat(this._retained), pattern)
}

MemoryPersistence.prototype.createRetainedStreamCombi = function (patterns) {
  return matchingStream([].concat(this._retained), patterns)
}

MemoryPersistence.prototype.addSubscriptions = function (client, subs, cb) {
  var stored = this._subscriptions.get(client.id)
  var trie = this._trie

  if (!stored) {
    stored = new Map()
    this._subscriptions.set(client.id, stored)
    this._clientsCount++
  }

  for (var i = 0; i < subs.length; i += 1) {
    var sub = subs[i]
    var qos = stored.get(sub.topic)
    var hasQoSGreaterThanZero = (qos !== undefined) && (qos > 0)
    if (sub.qos > 0) {
      trie.add(sub.topic, {
        clientId: client.id,
        topic: sub.topic,
        qos: sub.qos
      })
    } else if (hasQoSGreaterThanZero) {
      trie.remove(sub.topic, {
        clientId: client.id,
        topic: sub.topic
      })
    }
    stored.set(sub.topic, sub.qos)
  }

  cb(null, client)
}

MemoryPersistence.prototype.removeSubscriptions = function (client, subs, cb) {
  var stored = this._subscriptions.get(client.id)
  var trie = this._trie

  if (stored) {
    for (var i = 0; i < subs.length; i += 1) {
      var topic = subs[i]
      var qos = stored.get(topic)
      if (qos !== undefined) {
        if (qos > 0) {
          trie.remove(topic, { clientId: client.id, topic: topic })
        }
        stored.delete(topic)
      }
    }

    if (stored.size === 0) {
      this._clientsCount--
      this._subscriptions.delete(client.id)
    }
  }

  cb(null, client)
}

MemoryPersistence.prototype.subscriptionsByClient = function (client, cb) {
  var subs = null
  var stored = this._subscriptions.get(client.id)
  if (stored) {
    subs = []
    for (var topicAndQos of stored) {
      subs.push({ topic: topicAndQos[0], qos: topicAndQos[1] })
    }
  }
  cb(null, subs, client)
}

MemoryPersistence.prototype.countOffline = function (cb) {
  return cb(null, this._trie.subscriptionsCount, this._clientsCount)
}

MemoryPersistence.prototype.subscriptionsByTopic = function (pattern, cb) {
  cb(null, this._trie.match(pattern))
}

MemoryPersistence.prototype.cleanSubscriptions = function (client, cb) {
  var trie = this._trie
  var stored = this._subscriptions.get(client.id)

  if (stored) {
    for (var topicAndQos of stored) {
      if (topicAndQos[1] > 0) {
        var topic = topicAndQos[0]
        trie.remove(topic, { clientId: client.id, topic: topic })
      }
    }

    this._clientsCount--
    this._subscriptions.delete(client.id)
  }

  cb(null, client)
}

MemoryPersistence.prototype.outgoingEnqueue = function (sub, packet, cb) {
  _outgoingEnqueue.call(this, sub, packet)
  process.nextTick(cb)
}

MemoryPersistence.prototype.outgoingEnqueueCombi = function (subs, packet, cb) {
  for (var i = 0; i < subs.length; i++) {
    _outgoingEnqueue.call(this, subs[i], packet)
  }
  process.nextTick(cb)
}

function _outgoingEnqueue (sub, packet) {
  var id = sub.clientId
  var queue = this._outgoing[id] || []

  this._outgoing[id] = queue

  queue[queue.length] = new Packet(packet)
}

MemoryPersistence.prototype.outgoingUpdate = function (client, packet, cb) {
  var i
  var clientId = client.id
  var outgoing = this._outgoing[clientId] || []
  var temp

  this._outgoing[clientId] = outgoing

  for (i = 0; i < outgoing.length; i++) {
    temp = outgoing[i]
    if (temp.brokerId === packet.brokerId &&
      temp.brokerCounter === packet.brokerCounter) {
      temp.messageId = packet.messageId
      return cb(null, client, packet)
    } else if (temp.messageId === packet.messageId) {
      outgoing[i] = packet
      return cb(null, client, packet)
    }
  }

  cb(new Error('no such packet'), client, packet)
}

MemoryPersistence.prototype.outgoingClearMessageId = function (client, packet, cb) {
  var i
  var clientId = client.id
  var outgoing = this._outgoing[clientId] || []
  var temp

  this._outgoing[clientId] = outgoing

  for (i = 0; i < outgoing.length; i++) {
    temp = outgoing[i]
    if (temp.messageId === packet.messageId) {
      outgoing.splice(i, 1)
      return cb(null, temp)
    }
  }

  cb()
}

MemoryPersistence.prototype.outgoingStream = function (client) {
  var queue = [].concat(this._outgoing[client.id] || [])

  return from2.obj(function match (size, next) {
    var entry

    while ((entry = queue.shift()) != null) {
      setImmediate(next, null, entry)
      return
    }

    if (!entry) this.push(null)
  })
}

MemoryPersistence.prototype.incomingStorePacket = function (client, packet, cb) {
  var id = client.id
  var store = this._incoming[id] || {}

  this._incoming[id] = store

  store[packet.messageId] = new Packet(packet)
  store[packet.messageId].messageId = packet.messageId

  cb(null)
}

MemoryPersistence.prototype.incomingGetPacket = function (client, packet, cb) {
  var id = client.id
  var store = this._incoming[id] || {}
  var err = null

  this._incoming[id] = store

  if (!store[packet.messageId]) {
    err = new Error('no such packet')
  }

  cb(err, store[packet.messageId])
}

MemoryPersistence.prototype.incomingDelPacket = function (client, packet, cb) {
  var id = client.id
  var store = this._incoming[id] || {}
  var toDelete = store[packet.messageId]
  var err = null

  if (!toDelete) {
    err = new Error('no such packet')
  } else {
    delete store[packet.messageId]
  }

  cb(err)
}

MemoryPersistence.prototype.putWill = function (client, packet, cb) {
  packet.brokerId = this.broker.id
  packet.clientId = client.id
  this._wills[client.id] = packet
  cb(null, client)
}

MemoryPersistence.prototype.getWill = function (client, cb) {
  cb(null, this._wills[client.id], client)
}

MemoryPersistence.prototype.delWill = function (client, cb) {
  var will = this._wills[client.id]
  delete this._wills[client.id]
  cb(null, will, client)
}

MemoryPersistence.prototype.streamWill = function (brokers) {
  var clients = Object.keys(this._wills)
  var wills = this._wills
  brokers = brokers || {}
  return from2.obj(function match (size, next) {
    var entry

    while ((entry = clients.shift()) != null) {
      if (!brokers[wills[entry].brokerId]) {
        setImmediate(next, null, wills[entry])
        return
      }
    }

    if (!entry) {
      this.push(null)
    }
  })
}

MemoryPersistence.prototype.getClientList = function (topic) {
  var clientSubs = this._subscriptions
  var entries = clientSubs.entries(clientSubs)
  return from2.obj(function match (size, next) {
    var entry
    while (!(entry = entries.next()).done) {
      if (entry.value[1].has(topic)) {
        setImmediate(next, null, entry.value[0])
        return
      }
    }
    next(null, null)
  })
}

MemoryPersistence.prototype.destroy = function (cb) {
  this._retained = null
  if (cb) {
    cb(null)
  }
}

module.exports = MemoryPersistence
