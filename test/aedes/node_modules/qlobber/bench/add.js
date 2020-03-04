/*globals options: false */
/*jslint node: true */
"use strict";

var common = require('./common');

module.exports = function ()
{
    var matcher = new options.Matcher();

    for (var i = 0; i < 20; i += 1)
    {
        common.add_bindings(matcher);
    }
};

