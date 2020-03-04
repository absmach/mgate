/*globals options: false */
/*jslint node: true */
"use strict";

var qlobber = require('..'),
    MapValQlobber = require('./options/_mapval').MapValQlobber,
    common = require('./common');

var matcher_default = new qlobber.Qlobber();
common.add_bindings(matcher_default);

var matcher_dedup = new qlobber.QlobberDedup();
common.add_bindings(matcher_dedup);

var matcher_mapval = new MapValQlobber();
common.add_bindings(matcher_mapval);

module.exports = function ()
{
    switch (options.Matcher)
    {
        case qlobber.QlobberDedup:
            common.match_search(matcher_dedup);
            break;

        case MapValQlobber:
            common.match_search(matcher_mapval);
            break;

        default:
            common.match_search(matcher_default);
            break;
    }
};
