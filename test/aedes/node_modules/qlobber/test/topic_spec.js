/*globals rabbitmq_test_bindings : false,
          rabbitmq_bindings_to_remove : false,
          rabbitmq_expected_results_before_remove: false,
          rabbitmq_expected_results_after_remove : false,
          rabbitmq_expected_results_after_remove_all : false,
          rabbitmq_expected_results_after_clear : false,
          describe: false,
          beforeEach: false,
          it: false */
/*jslint node: true */
"use strict";

var expect = require('chai').expect,
    qlobber = require('..'),
    common = require('./common');

describe('qlobber', function ()
{
    var matcher;

    beforeEach(function (done)
    {
        matcher = new qlobber.Qlobber();
        done();
    });

    function add_bindings(bindings, mapper)
    {
        mapper = mapper || function (topic) { return topic; };

        bindings.forEach(function (topic_val)
        {
            matcher.add(topic_val[0], mapper(topic_val[1]));
        });
    }

    function remove_duplicates_filter(item, index, arr)
    {
        return item !== arr[index - 1];
    }

    Array.prototype.remove_duplicates = function ()
    {
        return this.sort().filter(remove_duplicates_filter);
    };

    function get_shortcuts(matcher)
    {
        var k, r = {};
        for (k of matcher._shortcuts.keys())
        {
            r[k] = matcher._shortcuts.get(k);
        }
        return r;
    }

    it('should support adding bindings', function ()
    {
        add_bindings(rabbitmq_test_bindings);
        expect(common.get_trie(matcher)).to.eql(common.expected_trie);
    });

    it('should pass rabbitmq test', function ()
    {
        add_bindings(rabbitmq_test_bindings);

        rabbitmq_expected_results_before_remove.forEach(function (test)
        {
            expect(matcher.match(test[0]).remove_duplicates(), test[0]).to.eql(test[1].sort());
            for (var v of test[1])
            {
                expect(matcher.test(test[0], v)).to.equal(true);
            }
            expect(matcher.test(test[0], 'xyzfoo')).to.equal(false);
        });
    });

    it('should support removing bindings', function ()
    {
        add_bindings(rabbitmq_test_bindings);

        rabbitmq_bindings_to_remove.forEach(function (i)
        {
            matcher.remove(rabbitmq_test_bindings[i-1][0],
                           rabbitmq_test_bindings[i-1][1]);
        });

        expect(common.get_trie(matcher)).to.eql({"a":{"b":{"c":{".":["t20"]},"b":{"c":{".":["t4"]},".":["t14"]},".":["t15"]},"*":{"c":{".":["t2"]},".":["t9"]},"#":{"b":{".":["t3"]},"#":{".":["t12"]}}},"#":{"#":{".":["t6"],"#":{".":["t24"]}},"b":{".":["t7"],"#":{".":["t26"]}},"*":{"#":{".":["t22"]}}},"*":{"*":{".":["t8"],"*":{".":["t18"]}},"b":{"c":{".":["t10"]}},"#":{"#":{".":["t23"]}},".":["t25"]},"b":{"b":{"c":{".":["t13"]}},"c":{".":["t16"]}},"":{".":["t17"]}});

        rabbitmq_expected_results_after_remove.forEach(function (test)
        {
            expect(matcher.match(test[0]).remove_duplicates(), test[0]).to.eql(test[1].sort());
            for (var v of test[1])
            {
                expect(matcher.test(test[0], v)).to.equal(true);
            }
            expect(matcher.test(test[0], 'xyzfoo')).to.equal(false);
        });
        
        /*jslint unparam: true */
        var remaining = rabbitmq_test_bindings.filter(function (topic_val, i)
        {
            return rabbitmq_bindings_to_remove.indexOf(i + 1) < 0;
        });
        /*jslint unparam: false */

        remaining.forEach(function (topic_val)
        {
            matcher.remove(topic_val[0], topic_val[1]);
        });
            
        expect(matcher.get_trie().size).to.equal(0);

        rabbitmq_expected_results_after_clear.forEach(function (test)
        {
            expect(matcher.match(test[0]).remove_duplicates(), test[0]).to.eql(test[1].sort());
            for (var v of test[1])
            {
                expect(matcher.test(test[0], v)).to.equal(true);
            }
            expect(matcher.test(test[0], 'xyzfoo')).to.equal(false);
        });
    });

    it('should support clearing the bindings', function ()
    {
        add_bindings(rabbitmq_test_bindings);

        matcher.clear();

        rabbitmq_expected_results_after_clear.forEach(function (test)
        {
            expect(matcher.match(test[0]).remove_duplicates(), test[0]).to.eql(test[1].sort());
            for (var v of test[1])
            {
                expect(matcher.test(test[0], v)).to.equal(true);
            }
            expect(matcher.test(test[0], 'xyzfoo')).to.equal(false);
        });
    });

    it('should support removing all values for a topic', function ()
    {
        add_bindings(rabbitmq_test_bindings);

        rabbitmq_bindings_to_remove.forEach(function (i)
        {
            matcher.remove(rabbitmq_test_bindings[i-1][0]);
        });
        
        expect(common.get_trie(matcher)).to.eql({"a":{"b":{"b":{"c":{".":["t4"]},".":["t14"]},".":["t15"]},"*":{"c":{".":["t2"]},".":["t9"]},"#":{"b":{".":["t3"]},"#":{".":["t12"]}}},"#":{"#":{".":["t6"],"#":{".":["t24"]}},"b":{".":["t7"],"#":{".":["t26"]}},"*":{"#":{".":["t22"]}}},"*":{"*":{".":["t8"],"*":{".":["t18"]}},"b":{"c":{".":["t10"]}},"#":{"#":{".":["t23"]}},".":["t25"]},"b":{"b":{"c":{".":["t13"]}},"c":{".":["t16"]}},"":{".":["t17"]}});

        rabbitmq_expected_results_after_remove_all.forEach(function (test)
        {
            expect(matcher.match(test[0]).remove_duplicates(), test[0]).to.eql(test[1].sort());
            for (var v of test[1])
            {
                expect(matcher.test(test[0], v)).to.equal(true);
            }
            expect(matcher.test(test[0], 'xyzfoo')).to.equal(false);
        });
    });

    it('should support functions as values', function ()
    {
        add_bindings(rabbitmq_test_bindings, function (topic)
        {
            return function ()
            {
                return topic;
            };
        });

        matcher.test_values = function (vals, val)
        {
            for (var v of vals)
            {
                if (v() === val)
                {
                    return true;
                }
            }

            return false;
        };

        matcher.equals_value = function (matched_value, test_value)
        {
            return matched_value() === test_value;
        };

        rabbitmq_expected_results_before_remove.forEach(function (test)
        {
            expect(matcher.match(test[0], test[0]).map(function (f)
            {
                return f();
            }).remove_duplicates()).to.eql(test[1].sort());
            for (var v of test[1])
            {
                expect(matcher.test(test[0], v)).to.equal(true);
            }
            expect(matcher.test(test[0], 'xyzfoo')).to.equal(false);
        });
    });

    it('should support undefined as a value', function ()
    {
        matcher.add('foo.bar');
        matcher.add('foo.*');
        expect(matcher.match('foo.bar')).to.eql([undefined, undefined]);
        expect(matcher.test('foo.bar')).to.equal(true);
    });

    it('should pass example in README', function ()
    {
        matcher.add('foo.*', 'it matched!');
        expect(matcher.match('foo.bar')).to.eql(['it matched!']);
        expect(matcher.test('foo.bar', 'it matched!')).to.equal(true);
    });

    it('should pass example in rabbitmq topic tutorial', function ()
    {
        matcher.add('*.orange.*', 'Q1');
        matcher.add('*.*.rabbit', 'Q2');
        matcher.add('lazy.#', 'Q2');
        expect(['quick.orange.rabbit',
                'lazy.orange.elephant',
                'quick.orange.fox',
                'lazy.brown.fox',
                'lazy.pink.rabbit',
                'quick.brown.fox',
                'orange',
                'quick.orange.male.rabbit',
                'lazy.orange.male.rabbit'].map(function (topic)
                {
                    return [matcher.match(topic).sort(),
                            matcher.test(topic, 'Q1'),
                            matcher.test(topic, 'Q2')];
                })).to.eql(
               [[['Q1', 'Q2'], true, true],
                [['Q1', 'Q2'], true, true],
                [['Q1'], true, false],
                [['Q2'], false, true],
                [['Q2', 'Q2'], false, true],
                [[], false, false],
                [[], false, false],
                [[], false, false],
                [['Q2'], false, true]]);
    });

    it('should not remove anything if not previously added', function ()
    {
        matcher.add('foo.*', 'it matched!');
        matcher.remove('foo');
        matcher.remove('foo.*', 'something');
        matcher.remove('bar.*');
        expect(matcher.match('foo.bar')).to.eql(['it matched!']);
        expect(matcher.test('foo.bar', 'it matched!')).to.equal(true);
    });

    it('should accept wildcards in match topics', function ()
    {
        matcher.add('foo.*', 'it matched!');
        matcher.add('foo.#', 'it matched too!');
        expect(matcher.match('foo.*').sort()).to.eql(['it matched too!', 'it matched!']);
        expect(matcher.match('foo.#').sort()).to.eql(['it matched too!', 'it matched!']);
        expect(matcher.test('foo.*', 'it matched!')).to.equal(true);
        expect(matcher.test('foo.*', 'it matched too!')).to.equal(true);
        expect(matcher.test('foo.#', 'it matched!')).to.equal(true);
        expect(matcher.test('foo.#', 'it matched too!')).to.equal(true);
    });

    it('should be configurable', function ()
    {
        matcher = new qlobber.Qlobber({
            separator: '/',
            wildcard_one: '+',
            wildcard_some: 'M'
        });

        matcher.add('foo/+', 'it matched!');
        matcher.add('foo/M', 'it matched too!');
        expect(matcher.match('foo/bar').sort()).to.eql(['it matched too!', 'it matched!']);
        expect(matcher.match('foo/bar/end').sort()).to.eql(['it matched too!']);
        expect(matcher.test('foo/bar', 'it matched!')).to.equal(true);
        expect(matcher.test('foo/bar', 'it matched too!')).to.equal(true);
        expect(matcher.test('foo/bar/end', 'it matched too!')).to.equal(true);
    });

    it('should match expected number of topics', function ()
    {
        // under coverage this takes longer
        this.timeout(60000);

        var i, j, vals;

        for (i = 0; i < 60000; i += 1)
        {
            for (j = 0; j < 5; j += 1)
            {
                matcher.add('app.test.user.behrad.testTopic-' + j, i);
            }
            matcher.add('app.test.user.behrad.*', i);
        }

        vals = matcher.match('app.test.user.behrad.testTopic-0');
        expect(vals.length).to.equal(120000);
        expect(vals.remove_duplicates().length).to.equal(60000);

        expect(matcher.test('app.test.user.behrad.testTopic-0', 0)).to.equal(true);
        expect(matcher.test('app.test.user.behrad.testTopic-0', 59999)).to.equal(true);
        expect(matcher.test('app.test.user.behrad.testTopic-0', 60000)).to.equal(false);
    });

    it('should visit trie', function ()
    {
        add_bindings(rabbitmq_test_bindings);

        let objs = [];

        for (let v of matcher.visit())
        {
            objs.push(v);
        }
        
        expect(objs).to.eql(common.expected_visits);
    });

    it('should restore trie', function ()
    {
        let restorer = matcher.get_restorer();

        for (let v of common.expected_visits)
        {
            restorer(v);
        }

        expect(common.get_trie(matcher)).to.eql(common.expected_trie);

        rabbitmq_expected_results_before_remove.forEach(function (test)
        {
            expect(matcher.match(test[0]).remove_duplicates(), test[0]).to.eql(test[1].sort());
            for (var v of test[1])
            {
                expect(matcher.test(test[0], v)).to.equal(true);
            }
            expect(matcher.test(test[0], 'xyzfoo')).to.equal(false);
        });
    });

    it('should not restore empty entries', function ()
    {
        matcher.add('foo.bar', 90);

        matcher._trie.get('foo').get('bar').get('.').shift();

        let objs = [];

        for (let v of matcher.visit())
        {
            objs.push(v);
        }

        expect(objs).to.eql([
            { type: 'start_entries' },
            { type: 'entry', i: 0, key: 'foo' },
            { type: 'start_entries' },
            { type: 'entry', i: 0, key: 'bar' },
            { type: 'start_entries' },
            { type: 'entry', i: 0, key: '.' },
            { type: 'start_values' },
            { type: 'end_values' },
            { type: 'end_entries' },
            { type: 'end_entries' },
            { type: 'end_entries' }
        ]);

        let matcher2 = new qlobber.Qlobber(),
            restorer = matcher2.get_restorer();

        for (let v of objs)
        {
            restorer(v);
        }

        expect(common.get_trie(matcher2)).to.eql({});
        expect(matcher2.match('foo.bar')).to.eql([]);
        expect(matcher2.test('foo.bar', 90)).to.equal(false);
    });

    it('should restore shortcuts', function ()
    {
        matcher = new qlobber.Qlobber({ cache_adds: true });
        add_bindings(rabbitmq_test_bindings);

        let shortcuts = get_shortcuts(matcher);
        expect(shortcuts).to.eql({
            'a.b.c': [ 't1', 't20' ],
            'a.*.c': [ 't2' ],
            'a.#.b': [ 't3' ],
            'a.b.b.c': [ 't4' ],
            '#': [ 't5' ],
            '#.#': [ 't6' ],
            '#.b': [ 't7' ],
            '*.*': [ 't8' ],
            'a.*': [ 't9' ],
            '*.b.c': [ 't10' ],
            'a.#': [ 't11' ],
            'a.#.#': [ 't12' ],
            'b.b.c': [ 't13' ],
            'a.b.b': [ 't14' ],
            'a.b': [ 't15' ],
            'b.c': [ 't16' ],
            '': [ 't17' ],
            '*.*.*': [ 't18' ],
            'vodka.martini': [ 't19' ],
            '*.#': [ 't21' ],
            '#.*.#': [ 't22' ],
            '*.#.#': [ 't23' ],
            '#.#.#': [ 't24' ],
            '*': [ 't25' ],
            '#.b.#': [ 't26' ]
        });

        let objs = [];
        for (let v of matcher.visit())
        {
            objs.push(v);
        }
        expect(objs).to.eql(common.expected_visits);

        let matcher2 = new qlobber.Qlobber({ cache_adds: true }),
            restorer = matcher2.get_restorer();
        for (let v of common.expected_visits)
        {
            restorer(v);
        }
        expect(get_shortcuts(matcher2)).to.eql({});

        matcher2 = new qlobber.Qlobber({ cache_adds: true });
        restorer = matcher2.get_restorer({ cache_adds: true });
        for (let v of common.expected_visits)
        {
            restorer(v);
        }
        expect(get_shortcuts(matcher2)).to.eql(shortcuts);
    });

    it('should add shortcuts to passed in Map', function ()
    {
        var topics = new Map();
        matcher = new qlobber.Qlobber({ cache_adds: topics });
        add_bindings(rabbitmq_test_bindings);
        var added = Array.from(topics.keys()).sort();
        var rtopics = new Set(rabbitmq_test_bindings.map(v => v[0]));
        expect(added).to.eql(Array.from(rtopics).sort());
    });
});
