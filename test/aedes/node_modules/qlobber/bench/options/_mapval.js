/*jslint node: true */

var util = require('util'),
    qlobber = require('../..'),
    Qlobber = qlobber.Qlobber;

function MapValQlobber(options)
{
    Qlobber.call(this, options);
}

util.inherits(MapValQlobber, Qlobber);

MapValQlobber.prototype._initial_value = function (val)
{
    return new Map().set(val, val);
};

MapValQlobber.prototype._add_value = function (vals, val)
{
    vals.set(val, val);
};

MapValQlobber.prototype._add_values = function (dest, origin)
{
    origin.forEach(function (val, key)
    {
        dest.set(key, val);
    });
};

MapValQlobber.prototype._remove_value = function (vals, val)
{
    vals.delete(val);
    return vals.size === 0;
};

MapValQlobber.prototype.test_values = function (vals, val)
{
    return vals.has(val);
};

MapValQlobber.prototype.match = function (topic)
{
    return this._match2(new Map(), topic);
};

exports.MapValQlobber = MapValQlobber;
