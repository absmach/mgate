# qlobber&nbsp;&nbsp;&nbsp;[![Build Status](https://travis-ci.org/davedoesdev/qlobber.png)](https://travis-ci.org/davedoesdev/qlobber) [![Coverage Status](https://coveralls.io/repos/davedoesdev/qlobber/badge.png?branch=master)](https://coveralls.io/r/davedoesdev/qlobber?branch=master) [![NPM version](https://badge.fury.io/js/qlobber.png)](http://badge.fury.io/js/qlobber)

Node.js globbing for amqp-like topics.

Example:

```javascript
var Qlobber = require('qlobber').Qlobber;
var matcher = new Qlobber();
matcher.add('foo.*', 'it matched!');
assert.deepEqual(matcher.match('foo.bar'), ['it matched!']);
assert(matcher.test('foo.bar', 'it matched!'));
```

The API is described [here](#tableofcontents).

qlobber is implemented using a trie, as described in the RabbitMQ blog posts [here](http://www.rabbitmq.com/blog/2010/09/14/very-fast-and-scalable-topic-routing-part-1/) and [here](http://www.rabbitmq.com/blog/2011/03/28/very-fast-and-scalable-topic-routing-part-2/).

## Installation

```shell
npm install qlobber
```

## Another Example

A more advanced example using topics from the [RabbitMQ topic tutorial](http://www.rabbitmq.com/tutorials/tutorial-five-python.html):

```javascript
var matcher = new Qlobber();
matcher.add('*.orange.*', 'Q1');
matcher.add('*.*.rabbit', 'Q2');
matcher.add('lazy.#', 'Q2');
assert.deepEqual(['quick.orange.rabbit',
                  'lazy.orange.elephant',
                  'quick.orange.fox',
                  'lazy.brown.fox',
                  'lazy.pink.rabbit',
                  'quick.brown.fox',
                  'orange',
                  'quick.orange.male.rabbit',
                  'lazy.orange.male.rabbit'].map(function (topic)
                  {
                      return matcher.match(topic).sort();
                  }),
                 [['Q1', 'Q2'],
                  ['Q1', 'Q2'],
                  ['Q1'],
                  ['Q2'],
                  ['Q2', 'Q2'],
                  [],
                  [],
                  [],
                  ['Q2']]);
```

## Licence

[MIT](LICENCE)

## Tests

qlobber passes the [RabbitMQ topic tests](https://github.com/rabbitmq/rabbitmq-server/blob/master/src/rabbit_tests.erl) (I converted them from Erlang to Javascript).

To run the tests:

```shell
grunt test
```

## Lint

```shell
grunt lint
```

## Code Coverage

```shell
grunt coverage
```

[Instanbul](http://gotwarlost.github.io/istanbul/) results are available [here](http://rawgit.davedoesdev.com/davedoesdev/qlobber/master/coverage/lcov-report/index.html).

Coveralls page is [here](https://coveralls.io/r/davedoesdev/qlobber).

## Benchmarks

```shell
grunt bench
```

qlobber is also benchmarked in [ascoltatori](https://github.com/mcollina/ascoltatori).

# API

_Source: [lib/qlobber.js](lib/qlobber.js)_

<a name="tableofcontents"></a>

- <a name="toc_qlobberoptions"></a>[Qlobber](#qlobberoptions)
- <a name="toc_qlobberprototypeaddtopic-val"></a><a name="toc_qlobberprototype"></a>[Qlobber.prototype.add](#qlobberprototypeaddtopic-val)
- <a name="toc_qlobberprototyperemovetopic-val"></a>[Qlobber.prototype.remove](#qlobberprototyperemovetopic-val)
- <a name="toc_qlobberprototypematchtopic"></a>[Qlobber.prototype.match](#qlobberprototypematchtopic)
- <a name="toc_qlobberprototypetesttopic-val"></a>[Qlobber.prototype.test](#qlobberprototypetesttopic-val)
- <a name="toc_qlobberprototypetest_valuesvals-val"></a>[Qlobber.prototype.test_values](#qlobberprototypetest_valuesvals-val)
- <a name="toc_qlobberprototypeclear"></a>[Qlobber.prototype.clear](#qlobberprototypeclear)
- <a name="toc_qlobberprototypevisit"></a>[Qlobber.prototype.visit](#qlobberprototypevisit)
- <a name="toc_qlobberprototypeget_restoreroptions"></a>[Qlobber.prototype.get_restorer](#qlobberprototypeget_restoreroptions)
- <a name="toc_qlobberdedupoptions"></a>[QlobberDedup](#qlobberdedupoptions)
- <a name="toc_qlobberdedupprototypetest_valuesvals-val"></a><a name="toc_qlobberdedupprototype"></a>[QlobberDedup.prototype.test_values](#qlobberdedupprototypetest_valuesvals-val)
- <a name="toc_qlobberdedupprototypematchtopic"></a>[QlobberDedup.prototype.match](#qlobberdedupprototypematchtopic)
- <a name="toc_qlobbertrueoptions"></a>[QlobberTrue](#qlobbertrueoptions)
- <a name="toc_qlobbertrueprototypetest_values"></a><a name="toc_qlobbertrueprototype"></a>[QlobberTrue.prototype.test_values](#qlobbertrueprototypetest_values)
- <a name="toc_qlobbertrueprototypematchtopic"></a>[QlobberTrue.prototype.match](#qlobbertrueprototypematchtopic)
- <a name="toc_visitorstreamqlobber"></a>[VisitorStream](#visitorstreamqlobber)
- <a name="toc_restorerstreamqlobber"></a>[RestorerStream](#restorerstreamqlobber)

## Qlobber([options])

> Creates a new qlobber.

**Parameters:**

- `{Object} [options]` Configures the qlobber. Use the following properties:
  - `{String} separator` The character to use for separating words in topics. Defaults to '.'. MQTT uses '/' as the separator, for example.

  - `{String} wildcard_one` The character to use for matching exactly one word in a topic. Defaults to '*'. MQTT uses '+', for example.

  - `{String} wildcard_some` The character to use for matching zero or more words in a topic. Defaults to '#'. MQTT uses '#' too.

  - `{Boolean|Map} cache_adds` Whether to cache topics when adding topic matchers. This will make adding multiple matchers for the same topic faster at the cost of extra memory usage. Defaults to `false`. If you supply a `Map` then it will be used to cache the topics (use this to enumerate all the topics in the qlobber).

<sub>Go: [TOC](#tableofcontents)</sub>

<a name="qlobberprototype"></a>

## Qlobber.prototype.add(topic, val)

> Add a topic matcher to the qlobber.

Note you can match more than one value against a topic by calling `add` multiple times with the same topic and different values.

**Parameters:**

- `{String} topic` The topic to match against.
- `{Any} val` The value to return if the topic is matched.

**Return:**

`{Qlobber}` The qlobber (for chaining).

<sub>Go: [TOC](#tableofcontents) | [Qlobber.prototype](#toc_qlobberprototype)</sub>

## Qlobber.prototype.remove(topic, [val])

> Remove a topic matcher from the qlobber.

**Parameters:**

- `{String} topic` The topic that's being matched against.
- `{Any} [val]` The value that's being matched. If you don't specify `val` then all matchers for `topic` are removed.

**Return:**

`{Qlobber}` The qlobber (for chaining).

<sub>Go: [TOC](#tableofcontents) | [Qlobber.prototype](#toc_qlobberprototype)</sub>

## Qlobber.prototype.match(topic)

> Match a topic.

**Parameters:**

- `{String} topic` The topic to match against.

**Return:**

`{Array}` List of values that matched the topic. This may contain duplicates. Use a [`QlobberDedup`](#qlobberdedupoptions) if you don't want duplicates.

<sub>Go: [TOC](#tableofcontents) | [Qlobber.prototype](#toc_qlobberprototype)</sub>

## Qlobber.prototype.test(topic, val)

> Test whether a topic match contains a value. Faster than calling [`match`](#qlobberprototypematchtopic) and searching the result for the value. Values are tested using [`test_values`](#qlobberprototypetest_valuesvals-val).

**Parameters:**

- `{String} topic` The topic to match against.
- `{Any} val` The value being tested for.

**Return:**

`{Boolean}` Whether matching against `topic` contains `val`.

<sub>Go: [TOC](#tableofcontents) | [Qlobber.prototype](#toc_qlobberprototype)</sub>

## Qlobber.prototype.test_values(vals, val)

> Test whether values found in a match contain a value passed to [`test`](#qlobberprototypetesttopic-val). You can override this to provide a custom implementation. Defaults to using `indexOf`.

**Parameters:**

- `{Array} vals` The values found while matching.
- `{Any} val` The value being tested for.

**Return:**

`{Boolean}` Whether `vals` contains `val`.

<sub>Go: [TOC](#tableofcontents) | [Qlobber.prototype](#toc_qlobberprototype)</sub>

## Qlobber.prototype.clear()

> Reset the qlobber.

Removes all topic matchers from the qlobber.

**Return:**

`{Qlobber}` The qlobber (for chaining).

<sub>Go: [TOC](#tableofcontents) | [Qlobber.prototype](#toc_qlobberprototype)</sub>

## Qlobber.prototype.visit()

> Visit each node in the qlobber's trie in turn.

**Return:**

`{Iterator}` An iterator on the trie. The iterator returns objects which, if fed (in the same order) to the function returned by [`get_restorer`](#qlobberprototypeget_restoreroptions) on a different qlobber, will build that qlobber's trie to the same state. The objects can be serialized using `JSON.stringify`, _if_ the values you store in the qlobber are also serializable.

<sub>Go: [TOC](#tableofcontents) | [Qlobber.prototype](#toc_qlobberprototype)</sub>

## Qlobber.prototype.get_restorer([options])

> Get a function which can restore the qlobber's trie to a state you retrieved
by calling [`visit`](#qlobberprototypevisit) on this or another qlobber.

**Parameters:**

- `{Object} [options]` Options for restoring the trie.
  - `{Boolean} cache_adds` Whether to cache topics when rebuilding the trie. This only applies if you also passed `cache_adds` as true in the [constructor](#qlobberoptions).

**Return:**

`{Function}` Function to call in order to rebuild the qlobber's trie. You should call this repeatedly with the objects you received from a call to [`visit`](#qlobberprototypevisit). If you serialized the objects, remember to deserialize them first (e.g. with `JSON.parse`)!

<sub>Go: [TOC](#tableofcontents) | [Qlobber.prototype](#toc_qlobberprototype)</sub>

## QlobberDedup([options])

> Creates a new de-duplicating qlobber.

Inherits from [`Qlobber`](#qlobberoptions).

**Parameters:**

- `{Object} [options]` Same options as [Qlobber](#qlobberoptions).

<sub>Go: [TOC](#tableofcontents)</sub>

<a name="qlobberdedupprototype"></a>

## QlobberDedup.prototype.test_values(vals, val)

> Test whether values found in a match contain a value passed to [`test`](#qlobberprototypetesttopic_val). You can override this to provide a custom implementation. Defaults to using `has`.

**Parameters:**

- `{Set} vals` The values found while matching ([ES6 Set](http://www.ecma-international.org/ecma-262/6.0/#sec-set-objects)).
- `{Any} val` The value being tested for.

**Return:**

`{Boolean}` Whether `vals` contains `val`.

<sub>Go: [TOC](#tableofcontents) | [QlobberDedup.prototype](#toc_qlobberdedupprototype)</sub>

## QlobberDedup.prototype.match(topic)

> Match a topic.

**Parameters:**

- `{String} topic` The topic to match against.

**Return:**

`{Set}` [ES6 Set](http://www.ecma-international.org/ecma-262/6.0/#sec-set-objects) of values that matched the topic.

<sub>Go: [TOC](#tableofcontents) | [QlobberDedup.prototype](#toc_qlobberdedupprototype)</sub>

## QlobberTrue([options])

> Creates a new qlobber which only stores the value `true`.

Whatever value you [`add`](#qlobberprototypeaddtopic-val) to this qlobber
(even `undefined`), a single, de-duplicated `true` will be stored. Use this
qlobber if you only need to test whether topics match, not about the values
they match to.

Inherits from [`Qlobber`](#qlobberoptions).

**Parameters:**

- `{Object} [options]` Same options as [Qlobber](#qlobberoptions).

<sub>Go: [TOC](#tableofcontents)</sub>

<a name="qlobbertrueprototype"></a>

## QlobberTrue.prototype.test_values()

> This override of [`test_values`](#qlobberprototypetest_valuesvals-val) always
returns `true`. When you call [`test`](#qlobberprototypetesttopic-val) on a
`QlobberTrue` instance, the value you pass is ignored since it only cares
whether a topic is matched.

**Return:**

`{Boolean}` Always `true`.

<sub>Go: [TOC](#tableofcontents) | [QlobberTrue.prototype](#toc_qlobbertrueprototype)</sub>

## QlobberTrue.prototype.match(topic)

> Match a topic.

Since `QlobberTrue` only cares whether a topic is matched and not about values
it matches to, this override of [`match`](#qlobberprototypematchtopic) just
calls [`test`](#qlobberprototypetesttopic-val) (with value `undefined`).

**Parameters:**

- `{String} topic` The topic to match against.

**Return:**

`{Boolean}` Whether the `QlobberTrue` instance matches the topic.

<sub>Go: [TOC](#tableofcontents) | [QlobberTrue.prototype](#toc_qlobbertrueprototype)</sub>

## VisitorStream(qlobber)

> Creates a new [`Readable`](https://nodejs.org/dist/latest-v8.x/docs/api/stream.html#stream_class_stream_readable) stream, in object mode, which calls [`visit`](#qlobberprototypevisit) on a qlobber to generate its data.

You could [`pipe`](https://nodejs.org/dist/latest-v8.x/docs/api/stream.html#stream_readable_pipe_destination_options) this to a [`JSONStream.stringify`](https://github.com/dominictarr/JSONStream#jsonstreamstringifyopen-sep-close) stream, for instance, to serialize the qlobber to JSON. See [this test](test/json.js#L14) for an example.

Inherits from [`Readable`](https://nodejs.org/dist/latest-v8.x/docs/api/stream.html#stream_class_stream_readable).

**Parameters:**

- `{Qlobber} qlobber` The qlobber to call [`visit`](#qlobberprototypevisit) on.

<sub>Go: [TOC](#tableofcontents)</sub>

## RestorerStream(qlobber)

> Creates a new [`Writable`](https://nodejs.org/dist/latest-v8.x/docs/api/stream.html#stream_class_stream_writable) stream, in object mode, which passes data written to it into the function returned by calling [`get_restorer`](#qlobberprototypeget_restoreroptions) on a qlobber.

You could [`pipe`](https://nodejs.org/dist/latest-v8.x/docs/api/stream.html#stream_readable_pipe_destination_options) a [`JSONStream.parse`](https://github.com/dominictarr/JSONStream#jsonstreamparsepath) stream to this, for instance, to deserialize the qlobber from JSON. See [this test](test/json.js#L33) for an example.

Inherits from [`Writable`](https://nodejs.org/dist/latest-v8.x/docs/api/stream.html#stream_class_stream_writable).

**Parameters:**

- `{Qlobber} qlobber` The qlobber to call [`get_restorer`](#qlobberprototypeget_restoreroptions) on.

<sub>Go: [TOC](#tableofcontents)</sub>

_&mdash;generated by [apidox](https://github.com/codeactual/apidox)&mdash;_
