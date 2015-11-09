var Error = React.createClass({
    render: function() {
        return (
            <div className="error-page">
                <div className="obj">
                    <i className="big text-danger fa fa-warning"></i>
                    <h1>Internal Melange Error</h1>
                    <br/>
                    <p>Error Message:</p>
                    <p className="error">Attempted to load route { this.props.route }, and could not find it.</p>
                    <br/>
                    <h3><a href="">
                        <i className="fa fa-fw fa-exclamation-circle"></i>
                        Report this issue.
                    </a></h3>
                </div>
            </div>
        )
    },
});

var NotFound = React.createClass({
    render: function() {
        return (
            <Error/>
        )
    },
})

module.exports = NotFound;
