var QlobberDedup = require('../..').QlobberDedup;
var times = require('./common');

function add(matcher, clientId, topic, qos)
{
    matcher.add(topic, clientId + ';' + topic + ';' + qos);
}

function remove(matcher, clientId, topic, qos)
{
    matcher.remove(topic, clientId + ';' + topic + ';' + qos);
}

function match(matcher, topic)
{
    return Array.from(matcher.match(topic)).map(function (m)
    {
        var parts = m.split(';');
        return {
            clientId: parts[0],
            topic: parts[1],
            qos: +parts[2]
        };
    });
}

function test(matcher, clientId, topic)
{
    var count = 0;
    var found = false;

    for (var m of matcher.match(topic))
    {
        var parts = m.split(';');
        if (parts[1] === topic)
        {
            count += 1;
            if (count > 1)
            {
                return false;
            }
            if (parts[0] === clientId)
            {
                found = true;
            }
        }
    }

    return (count === 1) && found;
}

times(QlobberDedup, add, remove, match, test);
