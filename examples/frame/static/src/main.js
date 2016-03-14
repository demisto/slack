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

  getFormattedTime(d) {
    let hours = d.getHours();
    let ap = 'AM';
    if (hours > 12) {
      hours = hours - 12;
      ap = 'PM';
    } else if (hours === 12) {
      ap = 'PM';
    } else if (hours === 0) {
      hours = 12;
    }
    const hoursStr = (hours < 10) ? '0' + hours : '' + hours;
    const min = d.getMinutes();
    const minStr = (min < 10) ? '0' + min : '' + min;
    return hoursStr + ':' + minStr + ' ' + ap;
  }

  render() {
    const {msg, colors} = this.props;
    const d = new Date(msg.ts * 1000);
    const t = this.getFormattedTime(d);
    return (
      <div style={{marginLeft: '30px', marginRight: '30px', width: '100%'}}>
        <div style={{color: colors.date, width: '8%', float: 'left'}}>{t}</div>
        <div style={{color: colors.channel, marginLeft: '15px', width: '15%', float: 'left'}}>#{msg.channel}</div>
        <div style={{color: colors.user, fontWeight: 'bold', marginLeft: '15px', width: '15%', float: 'left'}}>@{msg.user}</div>
        <div style={{color: colors.text, marginLeft: '15px', width: '55%', float: 'left'}}>{msg.text}</div>
      </div>
    );
  }
}

class DateSeparator extends Component {
  static propTypes = {
    msg: PropTypes.object,
    colors: PropTypes.object
  };

  constructor(props) {
    super(props);
  }

  render() {
    const {msg, colors} = this.props;
    const d = new Date(msg.ts * 1000);
    return (
      <div style={{color: colors.dateSep, backgroundColor: colors.dateSepBack, marginLeft: '30px', marginRight: '30px', height: '26px', textAlign: 'center', width: '100%', float: 'left'}}>
        {d.toDateString()}
      </div>
    );
  }
}

class Frame extends Component { // eslint-disable-line react/no-multi-comp
  constructor(props) {
    super(props);
    this.state = {messages: [], colors: {date: '#999', dateSep: '#fff', dateSepBack: '#608081', user: '#000', channel: '#000', text: '#000', background: '#fff'}};
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
    fetch('/hist', {
      method: 'GET',
      headers: {
        'Accept': 'application/json'
      }
    })
    .then(r => r.json())
    .then(r => {
      this.setState({messages: r});
    });
    ws.open(this.handler.bind(this));
  }

  componentDidUpdate() {
    window.scrollTo(0, document.body.scrollHeight);
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

  diffDate(ts1, ts2) {
    const d1 = new Date(ts1 * 1000);
    const d2 = new Date(ts2 * 1000);
    return d1.toDateString() !== d2.toDateString();
  }

  render() {
    const {messages, colors} = this.state;
    const msgObjects = [];
    for (let i = 0; i < messages.length; i++) {
      if (i === 0 || this.diffDate(messages[i-1].ts, messages[i].ts)) {
        msgObjects.push(<DateSeparator key={messages[i].channel + messages[i].ts + '_sep'} msg={messages[i]} colors={colors} />);
      }
      msgObjects.push(<Message key={messages[i].channel + messages[i].ts + '_' + i} msg={messages[i]} colors={colors} />);
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
