/*globals options: false */
/*jslint node: true */
"use strict";

var assert = require('assert'),
    qlobber = require('..'),
    MapValQlobber = require('./options/_mapval').MapValQlobber;

var matcher_options = {
    separator: "/",
    wildcard_one: "+",
    cache_adds: true
};

function add_bindings(matcher)
{
    var i, j;
    for (i = 0; i < 60000; i += 1)
    {
        for (j = 0; j < 100; j += 1)
        {
            matcher.add('app/test/user/behrad/testTopic-' + j, i);
        }
        matcher.add('app/test/user/behrad/+', i);
    }
}

var matcher_default = new qlobber.Qlobber(matcher_options);
add_bindings(matcher_default);

var matcher_dedup = new qlobber.QlobberDedup(matcher_options);
add_bindings(matcher_dedup);

var matcher_mapval = new MapValQlobber(matcher_options);
add_bindings(matcher_mapval);

module.exports = function ()
{
    var i, j, vals;

    // This test is too slow with 60000
    for (i = 0; i < 10; i += 1)
    {
        for (j = 0; j < 100; j += 1)
        {
            // Typically app would match and search each time
            switch (options.Matcher)
            {
                case qlobber.QlobberDedup:
                    vals = matcher_dedup.match('app/test/user/behrad/testTopic-' + j);
                    assert(vals.has(i));
                    break;

                case MapValQlobber:
                    vals = matcher_mapval.match('app/test/user/behrad/testTopic-' + j);
                    assert(vals.has(i));
                    break;

                default:
                    vals = matcher_default.match('app/test/user/behrad/testTopic-' + j);
                    assert(vals.indexOf(i) >= 0);
                    break;
            }
        }
    }
};
