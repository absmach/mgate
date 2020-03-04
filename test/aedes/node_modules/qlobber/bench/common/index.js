/*globals rabbitmq_test_bindings: false,
          rabbitmq_bindings_to_remove: false,
          rabbitmq_expected_results_before_remove: false,
          rabbitmq_expected_results_after_remove: false,
          options: false */
/*jslint node: true */
"use strict";

var assert = require('assert');
var qlobber = require('../..');
var MapValQlobber = require('../options/_mapval').MapValQlobber;
var expect = require('chai').expect;
require('../../test/rabbitmq.js');

function remove_duplicates_filter(item, index, arr)
{
    return item !== arr[index - 1];
}

function remove_duplicates(arr)
{
    return arr.sort().filter(remove_duplicates_filter);
}

exports.remove_duplicates = remove_duplicates;

exports.add_bindings = function(matcher)
{
    var i, topic_val;

    for (i = 0; i < rabbitmq_test_bindings.length; i += 1)
    {
        topic_val = rabbitmq_test_bindings[i];
        matcher.add(topic_val[0], topic_val[1]);
    }
};

exports.remove_bindings = function(matcher)
{
    var i, r, test, vals;

    for (i = 0; i < rabbitmq_bindings_to_remove.length; i += 1)
    {
        r = rabbitmq_test_bindings[rabbitmq_bindings_to_remove[i] - 1];
        matcher.remove(r[0], r[1]);
    }

    if (options.check)
    {
        for (i = 0; i < rabbitmq_expected_results_after_remove.length; i += 1)
        {
            test = rabbitmq_expected_results_after_remove[i];
            vals = matcher.match(test[0]);
            
            if (options.Matcher === qlobber.QlobberDedup)
            {
                vals = Array.from(vals).sort();
            }
            else if (options.Matcher === MapValQlobber)
            {
                vals = Array.from(vals.keys()).sort();
            }
            else
            {
                vals = remove_duplicates(vals);
            }

            expect(vals).to.eql(test[1].sort());
        }
    }
};

exports.match = function(matcher)
{
    var i, test, vals;

    for (i = 0; i < rabbitmq_expected_results_before_remove.length; i += 1)
    {
        test = rabbitmq_expected_results_before_remove[i];
        vals = matcher.match(test[0]);

        if (options.Matcher === qlobber.Qlobber)
        {
            vals = remove_duplicates(vals);
        }

        if (options.check)
        {
            if (options.Matcher === qlobber.QlobberDedup)
            {
                vals = Array.from(vals).sort();
            }
            else if (options.Matcher === MapValQlobber)
            {
                vals = Array.from(vals.keys()).sort();
            }

            expect(vals).to.eql(test[1].sort());
        }
    }
};

exports.match_search = function(matcher)
{
    var i, j, test, vals;

    for (i = 0; i < rabbitmq_expected_results_before_remove.length; i += 1)
    {
        test = rabbitmq_expected_results_before_remove[i];

        for (j = 0; j < test[1].length; j += 1)
        {
            // Typically app would match and search each time
            vals = matcher.match(test[0]);

            if ((options.Matcher === qlobber.QlobberDedup) ||
                (options.Matcher === MapValQlobber))
            {
                assert(vals.has(test[1][j]));
            }
            else
            {
                assert(vals.indexOf(test[1][j]) >= 0);
            }
        }
    }
};

exports.test = function(matcher)
{
    var i, j, test, vals;

    for (i = 0; i < rabbitmq_expected_results_before_remove.length; i += 1)
    {
        test = rabbitmq_expected_results_before_remove[i];

        for (j = 0; j < test[1].length; j += 1)
        {
            assert(matcher.test(test[0], test[1][j]));
        }
    }
};
