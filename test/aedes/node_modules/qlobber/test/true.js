/*globals rabbitmq_test_bindings: false,
          rabbitmq_expected_results_before_remove: false */
/*jshint node: true, mocha: true */
"use strict";

var expect = require('chai').expect,
    qlobber = require('..'),
    QlobberTrue = qlobber.QlobberTrue,
    expected_visits = require('./common').expected_true_visits;

describe('true', function ()
{
    it('should add and test', function ()
    {
        var matcher = new QlobberTrue();
        matcher.add('a.b.c.d');
        expect(matcher.test('a.b.c.d')).to.equal(true);
        expect(matcher.test('foo.bar')).to.equal(false);
    });

    it('should add only true when adding multiple values', function ()
    {
        var matcher = new QlobberTrue();
        matcher.add('a.b.c.d', 'foo');
        matcher.add('a.b.c.d', 'bar');
        expect(matcher.test('a.b.c.d')).to.equal(true);
        expect(matcher.test('foo.bar')).to.equal(false);
    });

    it('should be able to remove', function ()
    {
        var matcher = new QlobberTrue();
        matcher.add('a.b.c.d');
        expect(matcher.test('a.b.c.d')).to.equal(true);
        matcher.remove('a.b.c.d');
        expect(matcher.test('a.b.c.d')).to.equal(false);
    });

    it('match should call test', function ()
    {
        var matcher = new QlobberTrue();
        matcher.add('a.b.c.d');
        matcher.add('a.b.c.*');
        expect(matcher.match('a.b.c.d')).to.equal(true);
        expect(matcher.match('a.b.c.e')).to.equal(true);
        expect(matcher.match('a.b.e')).to.equal(false);
        matcher.remove('a.b.c.d');
        expect(matcher.match('a.b.c.d')).to.equal(true);
    });

    it('should work with shortcuts', function ()
    {
        var matcher = new QlobberTrue({ cache_adds: true });
        matcher.add('a.b.c.d');
        expect(matcher.test('a.b.c.d')).to.equal(true);
        matcher.remove('a.b.c.d', 'foobar');
        expect(matcher.test('a.b.c.d')).to.equal(false);
    });

    function add_bindings(matcher, bindings, mapper)
    {
        mapper = mapper || function (topic) { return topic; };

        bindings.forEach(function (topic_val)
        {
            matcher.add(topic_val[0], mapper(topic_val[1]));
        });
    }

    function get_trie(matcher, t)
    {
        t = t || matcher.get_trie();
        var k, r = {};
        for (k of t.keys())
        {
            if (k === '.')
            {
                r[k] = t.get(k);
            }
            else
            {
                r[k] = get_trie(matcher, t.get(k));
            }
        }
        return r;
    }

    it('should visit trie', function ()
    {
        let matcher = new QlobberTrue();
        add_bindings(matcher, rabbitmq_test_bindings);

        let objs = [];

        for (let v of matcher.visit())
        {
            objs.push(v);
        }

        expect(objs).to.eql(expected_visits);
    });

    it('should restore trie', function ()
    {
        let matcher = new QlobberTrue(),
            restorer = matcher.get_restorer();

        for (let v of expected_visits)
        {
            restorer(v);
        }

        expect(get_trie(matcher)).to.eql({"a":{"b":{"c":{".":true},"b":{"c":{".":true},".":true},".":true},"*":{"c":{".":true},".":true},"#":{"b":{".":true},".":true,"#":{".":true}}},"#":{".":true,"#":{".":true,"#":{".":true}},"b":{".":true,"#":{".":true}},"*":{"#":{".":true}}},"*":{"*":{".":true,"*":{".":true}},"b":{"c":{".":true}},"#":{".":true,"#":{".":true}},".":true},"b":{"b":{"c":{".":true}},"c":{".":true}},"":{".":true},"vodka":{"martini":{".":true}}});

        rabbitmq_expected_results_before_remove.forEach(function (test)
        {
            expect(matcher.match(test[0]), test[0]).to.equal(true);
            for (var v of test[1])
            {
                expect(matcher.test(test[0], v)).to.equal(true);
            }
            expect(matcher.test(test[0], 'xyzfoo')).to.equal(true);
        });

        expect(matcher.match('xyzfoo')).to.equal(true);
        expect(matcher.test('xyzfoo')).to.equal(true);
    });
});
