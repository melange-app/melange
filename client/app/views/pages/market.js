var S = require('../../store');

var TopApp = React.createClass({
    gotoApp: function() {
        var Routes = require('../../routes');

        S.goto(
            Routes.market,
            {
                id: "com.getmelange.plugin.chat",
            }
        );
    },
    render: function() {
        return (
            <div className="col-sm-6">
                <div onClick={ this.gotoApp } className="panel top-app">
                    <div className="app-icon" style={ getBackground() }></div>
                    <h3><span className="text-muted">{ this.props.rank }</span> Chat</h3>
                    <p>from Hunter Leath, com.getmelange.plugin.chat</p>
                </div>
            </div>
        )
    }
});

var MarketHome = React.createClass({
    gotoApp: function(app) {
        return function() {
            var Routes = require('../../routes');

            S.goto(
                Routes.market,
                {
                    id: "com.getmelange.plugin.chat",
                }
            );
        }
    },
    render: function() {
        var apps = [
            <TopApp rank="1"/>,
            <TopApp rank="2"/>,
            <TopApp rank="3"/>,
            <TopApp rank="4"/>,
        ];

        console.log(this.props);
        
        return (
            <div className="market">
                <div className="featured">
                    <div onClick={ this.gotoApp('com.getmelange.plugins.notes') }
                         className="app first" style={ getBackground("cover") }></div>
                    <div onClick={ this.gotoApp('com.getmelange.plugins.notes') }
                         className="app second" style={ getBackground("cover") }></div>
                    <div onClick={ this.gotoApp('com.getmelange.plugins.id') }
                         className="app third" style={ getBackground("cover") }></div>
                </div>
                <div className="content">
                    <div className="alert alert-warning update-alert">
                        You have <b>6</b> applications with updates available. Please review them now.
                    </div>
                    
                    <h3>Top Applications</h3>

                    <div className="row top">
                        { apps }
                    </div>
                </div>
            </div>
        )
    },
});

var MarketView = React.createClass({
    render: function() {
        return (
            <div className="market-view">
                <div className="row">
                    <div className="col-sm-4">
                        <div className="padding-container">
                            <div className="icon-container">
                                <div className="spacer"></div>
                                <div className="app-icon" style={ getBackground() }></div>
                            </div>
                        </div>

                        <div className="actions">
                            <div className="btn btn-block btn-large btn-success">
                                Install App
                            </div>

                            <div className="list-group info">
                                <div className="list-group-item">
                                    <span className="title text-muted">
                                        Author:
                                    </span>
                                    Hunter Leath
                                </div>
                                <div className="list-group-item">
                                    <span className="title text-muted">
                                        Version:
                                    </span>
                                    0.0.1
                                </div>
                                <div className="list-group-item">
                                    <span className="title text-muted">
                                        Last Updated:
                                    </span>
                                    February 15, 2015
                                </div>
                                <div className="list-group-item">
                                    <span className="title text-muted">
                                        Id:
                                    </span>
                                    { this.props.app }
                                </div>
                            </div>

                            <div className="row">
                                <div className="col-xs-6">
                                    <div className="btn btn-block btn-primary">
                                        Blockchain
                                    </div>
                                </div>
                                <div className="col-xs-6">
                                    <div className="btn btn-block btn-default">
                                        Report
                                    </div>
                                </div>
                            </div>
                        </div>
                    </div>
                    <div className="col-sm-8">
                        <h1>Melange Chat</h1>
                        <p className="text-muted">Hunter Leath</p>

                        <hr/>

                        <h3>App Description</h3>
                        <p>
                            Pellentesque dapibus suscipit ligula.
                            Donec posuere augue in quam.  Etiam vel
                            tortor sodales tellus ultricies commodo.
                            Suspendisse potenti.  Aenean in sem ac leo
                            mollis blandit.  Donec neque quam,
                            dignissim in, mollis nec, sagittis eu,
                            wisi.  Phasellus lacus.  Etiam laoreet
                            quam sed arcu.  Phasellus at dui in ligula
                            mollis ultricies.  Integer placerat
                            tristique nisl.  Praesent augue.  Fusce
                            commodo.  Vestibulum convallis, lorem a
                            tempus semper, dui dui euismod elit, vitae
                            placerat urna tortor vitae lacus.  Nullam
                            libero mauris, consequat quis, varius et,
                            dictum id, arcu.  Mauris mollis tincidunt
                            felis.  Aliquam feugiat tellus ut neque.
                            Nulla facilisis, risus a rhoncus
                            fermentum, tellus tellus lacinia purus, et
                            dictum nunc justo sit amet elit.
                        </p>
                    </div>
                </div>
            </div>
        )
    },
});

var Market = React.createClass({
    render: function() {
        if (this.props.data !== undefined) {
            if (this.props.data.has('id')) {
                return (
                    <MarketView app={ this.props.data.get('id') }/>
                );
            }
        }

        return (
            <MarketHome/>
        );
    },
});

module.exports = Market
