/*globals rabbitmq_test_bindings: false,
          rabbitmq_expected_results_before_remove: false */
/*jslint mocha: true */
"use strict";

var expect = require('chai').expect,
    util = require('util'),
    Qlobber = require('..').Qlobber;

function QlobberMatchCount (options)
{
    Qlobber.call(this, options);
}

util.inherits(QlobberMatchCount, Qlobber);

QlobberMatchCount.prototype._add_values = function (dest, origin, count)
{
    if (count)
    {
        dest[0] += origin.length;
    }
    else
    {
        Qlobber.prototype._add_values(dest, origin);
    }
};

QlobberMatchCount.prototype.count = function (topic)
{
    return this._match([0], 0, topic.split(this._separator), this._trie, true)[0];
};

describe('qlobber-match-count', function ()
{
    function remove_duplicates_filter(item, index, arr)
    {
        return item !== arr[index - 1];
    }

    Array.prototype.remove_duplicates = function ()
    {
        return this.sort().filter(remove_duplicates_filter);
    };

    it('should be able to count matches', function ()
    {
        var matcher = new QlobberMatchCount();

        rabbitmq_test_bindings.forEach(function (topic_val)
        {
            matcher.add(topic_val[0], topic_val[1]);
        });

        rabbitmq_expected_results_before_remove.forEach(function (test)
        {
			var matched = matcher.match(test[0]);
            expect(matched.remove_duplicates(), test[0]).to.eql(test[1].sort());
            expect(matcher.count(test[0])).to.equal(matched.length);
        });
    });
});
