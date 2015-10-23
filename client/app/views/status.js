var Components = require('../components');

var Status = Components.createStateful({
    stateName: "f",
    filterState: function(s) {
        return s.status;
    },
    render: function() {
        return (
            <div className="status">
                <MessageLoading loading={ this.state.f.get('statusLoading') }
                                text={ this.state.f.get('statusText') }/>
                <StatusLight color={ this.state.f.get('connectionColor') }
                             text={ this.state.f.get('connectionText') }/>
            </div>
        )
    },
});

var MessageLoading = React.createClass({
    render: function() {
        var display = "none";
        if (this.props.loading) {
            display = "block";
        }
        
        var style={
            "display": display,
        }
        
        return (
            <div>
                <div style={ style } className="pull-left indicator"></div>
                <div className="pull-left">
                    <p>{ this.props.text }</p>
                </div>
            </div>
        );
    }
});

var StatusLight = React.createClass({
    render: function() {
        var lightClass="light pull-right ";
        lightClass += this.props.color;
        
        return (
            <div>
                <div className="pull-right">
                    <p>
                        { this.props.text }
                    </p>
                </div>
                <div className={ lightClass }></div>
            </div>
        )
    }
});

module.exports = Status;
