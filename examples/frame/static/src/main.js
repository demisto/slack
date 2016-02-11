import 'babel-core/polyfill';
import React, {Component, PropTypes} from 'react';
import ReactDOM from 'react-dom';
import * as ws from './ws';

const MAX_MESSAGES = 1000;

class Message extends Component {
  static propTypes = {
    msg: PropTypes.object,
    colors: PropTypes.object
  };

  constructor(props) {
    super(props);
  }

  render() {
    const {msg, colors} = this.props;
    return (
      <div>
        <span style={{color: colors.date}}>{msg.ts}</span>&nbsp;
        <span style={{color: colors.channel}}>#{msg.channel}</span>&nbsp;
        <span style={{color: colors.user}}>@{msg.user}</span>&nbsp;
        <span style={{color: colors.text}}>{msg.text}</span>
      </div>
    );
  }
}

class Frame extends Component { // eslint-disable-line react/no-multi-comp
  constructor(props) {
    super(props);
    this.state = {messages: [], colors: {date: 'blue', user: 'red', channel: '#618081', text: 'black', background: 'white'}};
  }

  componentDidMount() {
    const {colors} = this.state;
    fetch('/state', {
      method: 'GET',
      headers: {
        'Accept': 'application/json'
      }
    })
    .then(r => r.json())
    .then(r => {
      this.setState({colors: Object.assign({}, colors, r)});
    });
    ws.open(this.handler.bind(this));
  }

  handler(msg) {
    const {messages} = this.state;
    const newMessages = Array.from(messages);
    newMessages.push(msg);
    if (newMessages.length > MAX_MESSAGES) {
      newMessages.shift();
    }
    this.setState({messages: newMessages});
  }

  render() {
    const {messages, colors} = this.state;
    const msgObjects = [];
    for (let i = 0; i < messages.length; i++) {
      msgObjects.push(<Message key={messages[i].channel + messages[i].ts} msg={messages[i]} colors={colors} />);
    }
    return (
      <div style={{backgroundColor: colors.background}}>
        {msgObjects}
      </div>
    );
  }
}

ReactDOM.render(
  <div>
    <Frame />
  </div>,
  document.getElementById('app')
);
