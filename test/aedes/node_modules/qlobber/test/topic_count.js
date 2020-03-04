/*globals rabbitmq_test_bindings: false,
          rabbitmq_bindings_to_remove: false */
/*jslint mocha: true */
"use strict";

var expect = require('chai').expect,
    util = require('util'),
    Qlobber = require('..').Qlobber;

function QlobberTopicCount (options)
{
    Qlobber.call(this, options);
    this.topic_count = 0;
}

util.inherits(QlobberTopicCount, Qlobber);

QlobberTopicCount.prototype._initial_value = function (val)
{
    this.topic_count += 1;
    return Qlobber.prototype._initial_value(val);
};

QlobberTopicCount.prototype._remove_value = function (vals, val)
{
    var removed = Qlobber.prototype._remove_value(vals, val);
    if (removed)
    {
        this.topic_count -= 1;
    }
    return removed;
};

QlobberTopicCount.prototype.clear = function ()
{
    this.topic_count = 0;
    return Qlobber.prototype.clear.call(this);
};

describe('qlobber-topic-count', function ()
{
    it('should be able to count topics added', function ()
    {
        var matcher = new QlobberTopicCount();

        rabbitmq_test_bindings.forEach(function (topic_val)
        {
            matcher.add(topic_val[0], topic_val[1]);
        });
        expect(matcher.topic_count).to.equal(25);

        rabbitmq_bindings_to_remove.forEach(function (i)
        {
            matcher.remove(rabbitmq_test_bindings[i-1][0],
                           rabbitmq_test_bindings[i-1][1]);
        });
        expect(matcher.topic_count).to.equal(21);

        matcher.clear();
        expect(matcher.topic_count).to.equal(0);
        expect(matcher.match('a.b.c').length).to.equal(0);
    });

    it('should not decrement count if entry does not exist', function ()
    {
        var matcher = new QlobberTopicCount();
        expect(matcher.topic_count).to.equal(0);

        matcher.add('foo.bar', 23);
        expect(matcher.topic_count).to.equal(1);

        matcher.remove('foo.bar', 24);
        expect(matcher.topic_count).to.equal(1);

        matcher.remove('foo.bar2', 23);
        expect(matcher.topic_count).to.equal(1);

        matcher.remove('foo.bar', 23);
        expect(matcher.topic_count).to.equal(0);

        matcher.remove('foo.bar', 24);
        expect(matcher.topic_count).to.equal(0);

        matcher.remove('foo.bar2', 23);
        expect(matcher.topic_count).to.equal(0);
    });
});
