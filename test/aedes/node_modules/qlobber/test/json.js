/*jshint node: true, mocha: true */
"use strict";

let expect = require('chai').expect,
    qlobber = require('..'),
    JSONStream = require('JSONStream'),
    streamBuffers = require('stream-buffers'),
    common = require('./common'),
    rabbitmq = require('./rabbitmq'),
    expected_json = JSON.stringify(common.expected_visits);

describe('json', function ()
{
    it('visit should support serializing to JSON', function (cb)
    {
        let matcher = new qlobber.QlobberDedup();

        for (let [topic, val] of rabbitmq.test_bindings)
        {
            matcher.add(topic, val);
        }

        new qlobber.VisitorStream(matcher)
            .pipe(JSONStream.stringify('[', ',', ']'))
            .pipe(new streamBuffers.WritableStreamBuffer())
            .on('finish', function ()
            {
                expect(this.getContentsAsString()).to.equal(expected_json);
                cb();
            });
    });

    it('restore should support deserializing from JSON', function (cb)
    {
        let matcher = new qlobber.QlobberDedup(),
            buf_stream = new streamBuffers.ReadableStreamBuffer();

        buf_stream.put(expected_json);
        buf_stream.stop();
        buf_stream
            .pipe(JSONStream.parse('*'))
            .pipe(new qlobber.RestorerStream(matcher)).on('finish', function ()
            {
                expect(common.get_trie(matcher)).to.eql(common.expected_trie);
                cb();
            });
    });
});

