var assert = require('assert');

var QlobberOpts = {
    wildcard_one: '+',
    wildcard_some: '#',
    separator: '/'
};

function time(f, matcher, arg)
{
    var start_time = new Date();
    f(matcher, arg);
    var end_time = new Date();
    console.log(f.name + ':', (end_time.getTime() - start_time.getTime()) + 'ms');
}

function add_to_qlobber(matcher, add)
{
    for (var i = 0; i < 300000; i += 1)
    {
        var x = Math.floor(Math.random() * 3 + 1);
        var clientId = 'someClientId/' + i;
        add(matcher, clientId, 'a/b/c/d/' + i, 1);
        add(matcher, clientId, 'a/b/c/d/def' + i, 1);
        add(matcher, clientId, 'a/b/c/d/id' + i, 1);
        add(matcher, clientId, 'a/b/c/public/test', 1);
        add(matcher, clientId, 'a/b/c/public/all', 1);
        add(matcher, clientId, 'a/b/c/public/' + x, 1);
    }
}

function remove_from_qlobber(matcher, remove)
{
    for (var i = 0; i < 10; i += 1)
    {
        var x = Math.floor(Math.random() * 3 + 1);
        var clientId = 'someClientId/' + i;
        remove(matcher, clientId, 'a/b/c/d/' + i, 1);
        remove(matcher, clientId, 'a/b/c/d/def' + i, 1);
        remove(matcher, clientId, 'a/b/c/d/id' + i, 1);
        remove(matcher, clientId, 'a/b/c/public/test', 1);
        remove(matcher, clientId, 'a/b/c/public/all', 1);
        remove(matcher, clientId, 'a/b/c/public/' + x, 1);
    }
}

function match_client_topics(matcher, match)
{
    for (var i = 0; i < 100000; i += 1)
    {
        match(matcher, 'a/b/c/d/' + i);
    } 
}

function match_public_topics(matcher, match)
{
    for (var i = 0; i < 10; i += 1)
    {
        assert.strictEqual(match(matcher, 'a/b/c/public/test').length, 299990);
    }
}

function test_public_topics(matcher, test)
{
    for (var i = 0; i < 300000; i += 1)
    {
        var clientId = 'someClientId/' + i;
        assert.strictEqual(test(matcher, clientId, 'a/b/c/public/test'), false);
    }
}

function times(QlobberClass, add, remove, match, test)
{
    gc();
    var start_mem = process.memoryUsage();

    var matcher = new QlobberClass(QlobberOpts);
    time(add_to_qlobber, matcher, add);
    time(add_to_qlobber, matcher, add);
    time(remove_from_qlobber, matcher, remove);
    time(match_client_topics, matcher, match);
    time(match_public_topics, matcher, match);
    time(test_public_topics, matcher, test);

    gc();
    var end_mem = process.memoryUsage();

    console.log(
        'heap:', ((end_mem.heapUsed - start_mem.heapUsed) / 1024 / 1024).toFixed(1) + 'MiB',
        'rss:', ((end_mem.rss - start_mem.rss) / 1024 / 1024).toFixed(1) + 'MiB');

    matcher.clear(); // ensure matcher is kept alive for gc above
}

module.exports = times;
