/*jshint node: true, mocha: true */
"use strict";

var expect = require('chai').expect,
    QlobberSub = require('../aedes/qlobber-sub');

describe('qlobber-sub', function ()
{
    it('should add and match a single value', function ()
    {
        var matcher = new QlobberSub();
        expect(matcher.subscriptionsCount).to.equal(0);
        matcher.add('foo.bar',
        {
            clientId: 'test1',
            topic: 'foo.bar',
            qos: 1
        });
        expect(matcher.subscriptionsCount).to.equal(1);
        expect(matcher.match('foo.bar')).to.eql([
        {
            clientId: 'test1',
            topic: 'foo.bar',
            qos: 1
        }]);
        expect(matcher.test('foo.bar',
        {
            clientId: 'test1',
            topic: 'foo.bar'
        })).to.equal(true);
        expect(matcher.test('foo.bar',
        {
            clientId: 'test2',
            topic: 'foo.bar'
        })).to.equal(false);
    });

    it('should dedup multiple values with same client ID and topic', function ()
    {
        var matcher = new QlobberSub();
        expect(matcher.subscriptionsCount).to.equal(0);
        matcher.add('foo.bar',
        {
            clientId: 'test1',
            topic: 'foo.bar',
            qos: 1
        });
        expect(matcher.subscriptionsCount).to.equal(1);
        matcher.add('foo.bar',
        {
            clientId: 'test1',
            topic: 'foo.bar',
            qos: 2
        });
        expect(matcher.subscriptionsCount).to.equal(1);
        expect(matcher.match('foo.bar')).to.eql([
        {
            clientId: 'test1',
            topic: 'foo.bar',
            qos: 2
        }]);
        expect(matcher.test('foo.bar',
        {
            clientId: 'test1',
            topic: 'foo.bar'
        })).to.equal(true);
        expect(matcher.test('foo.bar',
        {
            clientId: 'test2',
            topic: 'foo.bar'
        })).to.equal(false);
    });

    it('should not dedup multiple values with different client IDs and same topic', function ()
    {
        var matcher = new QlobberSub();
        expect(matcher.subscriptionsCount).to.equal(0);
        matcher.add('foo.bar',
        {
            clientId: 'test1',
            topic: 'foo.bar',
            qos: 1
        });
        expect(matcher.subscriptionsCount).to.equal(1);
        matcher.add('foo.bar',
        {
            clientId: 'test2',
            topic: 'foo.bar',
            qos: 2
        });
        expect(matcher.subscriptionsCount).to.equal(2);
        expect(matcher.match('foo.bar')).to.eql([
        {
            clientId: 'test1',
            topic: 'foo.bar',
            qos: 1
        },
        {
            clientId: 'test2',
            topic: 'foo.bar',
            qos: 2
        }]);
        expect(matcher.test('foo.bar',
        {
            clientId: 'test1',
            topic: 'foo.bar'
        })).to.equal(false);
        expect(matcher.test('foo.bar',
        {
            clientId: 'test2',
            topic: 'foo.bar'
        })).to.equal(false);
    });

    it('should not dedup multiple values with same client ID and different topics', function ()
    {
        var matcher = new QlobberSub();
        expect(matcher.subscriptionsCount).to.equal(0);
        matcher.add('foo.bar',
        {
            clientId: 'test1',
            topic: 'foo.bar',
            qos: 1
        });
        expect(matcher.subscriptionsCount).to.equal(1);
        matcher.add('foo.*',
        {
            clientId: 'test1',
            topic: 'foo.*',
            qos: 2
        });
        expect(matcher.subscriptionsCount).to.equal(2);
        expect(matcher.match('foo.bar')).to.eql([
        {
            clientId: 'test1',
            topic: 'foo.bar',
            qos: 1
        },
        {
            clientId: 'test1',
            topic: 'foo.*',
            qos: 2
        }]);
        expect(matcher.test('foo.bar',
        {
            clientId: 'test1',
            topic: 'foo.bar'
        })).to.equal(true);
        expect(matcher.test('foo.bar',
        {
            clientId: 'test1',
            topic: 'foo.*'
        })).to.equal(true);
        expect(matcher.test('foo.bar',
        {
            clientId: 'test1',
            topic: 'bar.*'
        })).to.equal(false);
        expect(matcher.test('foo.bar',
        {
            clientId: 'test2',
            topic: 'foo.bar'
        })).to.equal(false);
    });

    it('should remove value', function ()
    {
        var matcher = new QlobberSub();
        expect(matcher.subscriptionsCount).to.equal(0);
        matcher.add('foo.bar',
        {
            clientId: 'test1',
            topic: 'foo.bar',
            qos: 1
        });
        expect(matcher.subscriptionsCount).to.equal(1);
        matcher.add('foo.bar',
        {
            clientId: 'test2',
            topic: 'foo.bar',
            qos: 2
        });
        expect(matcher.subscriptionsCount).to.equal(2);
        matcher.remove('foo.bar',
        {
            clientId: 'test1',
            topic: 'foo.bar'
        });
        expect(matcher.subscriptionsCount).to.equal(1);
        expect(matcher.match('foo.bar')).to.eql([
        {
            clientId: 'test2',
            topic: 'foo.bar',
            qos: 2
        }]);
        expect(matcher.test('foo.bar',
        {
            clientId: 'test1',
            topic: 'foo.bar'
        })).to.equal(false);
        expect(matcher.test('foo.bar',
        {
            clientId: 'test2',
            topic: 'foo.bar'
        })).to.equal(true);
    });

    it('should be able to pass specific topic to match', function ()
    {
        var matcher = new QlobberSub();
        expect(matcher.subscriptionsCount).to.equal(0);
        matcher.add('foo.bar',
        {
            clientId: 'test1',
            topic: 'foo.bar',
            qos: 1
        });
        expect(matcher.subscriptionsCount).to.equal(1);
        matcher.add('foo.*',
        {
            clientId: 'test1',
            topic: 'foo.*',
            qos: 1
        });
        expect(matcher.subscriptionsCount).to.equal(2);
        matcher.add('foo.bar',
        {
            clientId: 'test2',
            topic: 'foo.bar',
            qos: 2
        });
        expect(matcher.subscriptionsCount).to.equal(3);
        expect(matcher.match('foo.bar')).to.eql([
        {
            clientId: 'test1',
            topic: 'foo.bar',
            qos: 1
        },
        {
            clientId: 'test2',
            topic: 'foo.bar',
            qos: 2
        },
        {
            clientId: 'test1',
            topic: 'foo.*',
            qos: 1
        }]);
        expect(matcher.match('foo.bar', 'foo.bar')).to.eql([
        {
            clientId: 'test1',
            qos: 1
        },
        {
            clientId: 'test2',
            qos: 2
        }]);
        expect(matcher.match('foo.bar', 'foo.*')).to.eql([
        {
            clientId: 'test1',
            qos: 1
        }]);
    });

    it("removing value shouldn't care about topic in value", function ()
    {
        var matcher = new QlobberSub();
        expect(matcher.subscriptionsCount).to.equal(0);
        matcher.add('foo.bar',
        {
            clientId: 'test1',
            topic: 'foo.bar',
            qos: 1
        });
        expect(matcher.subscriptionsCount).to.equal(1);
        matcher.add('foo.bar',
        {
            clientId: 'test2',
            topic: 'foo.bar',
            qos: 2
        });
        expect(matcher.subscriptionsCount).to.equal(2);
        matcher.remove('foo.bar',
        {
            clientId: 'test1',
            topic: 'foo.bar2'
        });
        expect(matcher.subscriptionsCount).to.equal(1);
        matcher.remove('foo.bar',
        {
            clientId: 'test3',
            topic: 'foo.bar'
        });
        expect(matcher.subscriptionsCount).to.equal(1);
        expect(matcher.match('foo.bar')).to.eql([
        {
            clientId: 'test2',
            topic: 'foo.bar',
            qos: 2
        }]);
        expect(matcher.test('foo.bar',
        {
            clientId: 'test1',
            topic: 'foo.bar'
        })).to.equal(false);
        expect(matcher.test('foo.bar',
        {
            clientId: 'test2',
            topic: 'foo.bar'
        })).to.equal(true);
        matcher.remove('foo.bar',
        {
            clientId: 'test2',
            topic: 'foo.bar2'
        });
        expect(matcher.subscriptionsCount).to.equal(0);
        matcher.remove('foo.bar',
        {
            clientId: 'test2',
            topic: 'foo.bar2'
        });
        expect(matcher.subscriptionsCount).to.equal(0);
    });

    it('should clear matcher', function ()
    {
        var matcher = new QlobberSub();
        expect(matcher.subscriptionsCount).to.equal(0);
        matcher.add('foo.bar',
        {
            clientId: 'test1',
            topic: 'foo.bar',
            qos: 1
        });
        expect(matcher.subscriptionsCount).to.equal(1);
        expect(matcher.match('foo.bar')).to.eql([
        {
            clientId: 'test1',
            topic: 'foo.bar',
            qos: 1
        }]);
        expect(matcher.test('foo.bar',
        {
            clientId: 'test1',
            topic: 'foo.bar'
        })).to.equal(true);
        expect(matcher.test('foo.bar',
        {
            clientId: 'test2',
            topic: 'foo.bar'
        })).to.equal(false);
        matcher.clear();
        expect(matcher.subscriptionsCount).to.equal(0);
        expect(matcher.match('foo.bar')).to.eql([]);
        expect(matcher.test('foo.bar',
        {
            clientId: 'test1',
            topic: 'foo.bar'
        })).to.equal(false);
        expect(matcher.test('foo.bar',
        {
            clientId: 'test2',
            topic: 'foo.bar'
        })).to.equal(false);
    });

    it('should count client subscription if has an existing subscription and is then added to a topic which already has a subscription for another client', function ()
    {
        var matcher = new QlobberSub();
        expect(matcher.subscriptionsCount).to.equal(0);
        matcher.add('foo.bar',
        {
            clientId: 'test1',
            topic: 'foo.bar',
            qos: 1
        });
        expect(matcher.subscriptionsCount).to.equal(1);
        matcher.add('foo.*',
        {
            clientId: 'test2',
            topic: 'foo.bar',
            qos: 1
        });
        expect(matcher.subscriptionsCount).to.equal(2);
        matcher.add('foo.*',
        {
            clientId: 'test1',
            topic: 'foo.bar',
            qos: 1
        });
        expect(matcher.subscriptionsCount).to.equal(3);
        matcher.remove('foo.*',
        {
            clientId: 'test1',
            topic: 'foo.bar',
        });
        expect(matcher.subscriptionsCount).to.equal(2);
    });
});
