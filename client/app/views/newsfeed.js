var Components = require('../components');
var Message = require("./message");
var Plugin = require('./plugin');

var NewsFeed = Components.createStateful({
    getInitialState: function() {
        return {
            numStories: 25,
        }
    },
    stateName: "f",
    filterState: function(s) {
        return {
            messages: s.messages,
            plugins: s.plugins,
        };
    },
    render: function() {
        var plugins = this.state.f.plugins;
        var stories = this.state.f.messages.store.filter(function(value) {
            return Plugin.isViewable(value, plugins);
        }).take(this.state.numStories).map(function(value) {
            return (
                <Message message={ value }/>
            );
        });
        
        return (
            <div className="newsfeed-holder">
                <div className="newsfeed">
                    <div className="header">
                        <div className="pull-right">
                            <a href="">
                                <i className="fa fa-arrows-alt"></i>
                            </a>
                        </div>
                        
                        Newsfeed
                    </div>

                    { stories }
                </div>
            </div>
        );
    }
});

module.exports = NewsFeed;
