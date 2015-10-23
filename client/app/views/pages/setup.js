var Setup = React.createClass({
    render: function() {
        return (
            <div className="setup">
                <p className="lead">Welcome to</p>
                
                <br/>
                
                <div className="row">
                    <div className="col-xs-3">
                        <img className="img-responsive" src="/img/icon.png"/>
                    </div>
                    <div className="col-xs-8">
                        <h1 className="style">Melange</h1>
                        <h3>Developer Preview</h3>
                    </div>
                </div>
                
                <br/>
                
                <div className="row">
                    <div className="col-xs-6">
                        <p className="lead">
                            The secure social network that gives you
                            back control of your data.
                        </p>

                        <div className="row text-center">
                            <div className="col-xs-6">
                                <h1><i className="fa fa-cloud-download"></i></h1>
                                <p>Install Apps</p>
                            </div>
                            <div className="col-xs-6">
                                <h1><i className="fa fa-users"></i></h1>
                                <p>Find Friends</p>
                            </div>
                        </div>
                    </div>
                    <div className="col-xs-6">
                        <SetupButton text="Sign Up"/>
                        <SetupButton text="Link Existing Account" color="pink"/>
                    </div>
                </div>
            </div>
        )
    },
});

var SetupButton = React.createClass({
    render: function() {
        var style = {};
        if (this.props.color != undefined) {
            style["border-bottom-color"] = this.props.color;
        }
        
        return (
            <div className="button" style={style}>
                <p>{ this.props.text } &rarr;</p>
            </div>
        );
    },
})

module.exports = Setup;
