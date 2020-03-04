/*jshint node: true, mocha: true */
"use strict";

var expect = require('chai').expect,
    qlobber = require('..'),
    QlobberDedup = qlobber.QlobberDedup;

describe('shortcut', function ()
{
    it('should add shortcut when adding', function ()
    {
        var matcher = new QlobberDedup({ cache_adds: true });
        expect(matcher._shortcuts.size).to.equal(0);
        matcher.add('a.b.c.d', 90);
        expect(matcher._shortcuts.size).to.equal(1);
        expect(matcher._shortcuts.get('a.b.c.d').size).to.equal(1);
        expect(Array.from(matcher.match('a.b.c.d'))).to.eql([90]);
        expect(matcher.test('a.b.c.d', 90)).to.equal(true);
    });

    it('should accept Map as cache', function ()
    {
        var topics = new Map();
        var matcher = new QlobberDedup({ cache_adds: topics });
        matcher.add('a.b.c.d', 90);
        expect(topics.size).to.equal(1);
        expect(topics.size).to.equal(1);
        expect(topics.get('a.b.c.d').size).to.equal(1);
        expect(Array.from(matcher.match('a.b.c.d'))).to.eql([90]);
        expect(matcher.test('a.b.c.d', 90)).to.equal(true);
    });

    it('should use shortcut when adding again', function ()
    {
        var matcher = new QlobberDedup({ cache_adds: true });
        expect(matcher._shortcuts.size).to.equal(0);
        matcher.add('a.b.c.d', 90);
        matcher.add('a.b.c.d', 91);
        expect(matcher._shortcuts.size).to.equal(1);
        expect(matcher._shortcuts.get('a.b.c.d').size).to.equal(2);
        expect(Array.from(matcher.match('a.b.c.d')).sort()).to.eql([90, 91]);
        expect(matcher.test('a.b.c.d', 90)).to.equal(true);
        expect(matcher.test('a.b.c.d', 91)).to.equal(true);
    });

    it('should remove shortcut when removing', function ()
    {
        var matcher = new QlobberDedup({ cache_adds: true });
        expect(matcher._shortcuts.size).to.equal(0);
        matcher.add('a.b.c.d', 90);
        matcher.remove('a.b.c.d', 90);
        expect(matcher._shortcuts.size).to.equal(0);
        expect(Array.from(matcher.match('a.b.c.d'))).to.eql([]);
        expect(matcher.test('a.b.c.d', 90)).to.equal(false);
    });

    it('should remove shortcut when removing all', function ()
    {
        var matcher = new QlobberDedup({ cache_adds: true });
        expect(matcher._shortcuts.size).to.equal(0);
        matcher.add('a.b.c.d', 90);
        matcher.remove('a.b.c.d');
        expect(matcher._shortcuts.size).to.equal(0);
        expect(Array.from(matcher.match('a.b.c.d'))).to.eql([]);
        expect(matcher.test('a.b.c.d', 90)).to.equal(false);
    });

    it('should clear shortcuts when matcher is cleared', function ()
    {
        var matcher = new QlobberDedup({ cache_adds: true });
        expect(matcher._shortcuts.size).to.equal(0);
        matcher.add('a.b.c.d', 90);
        matcher.clear();
        expect(matcher._shortcuts.size).to.equal(0);
        expect(Array.from(matcher.match('a.b.c.d'))).to.eql([]);
        expect(matcher.test('a.b.c.d', 90)).to.equal(false);
    });
});
