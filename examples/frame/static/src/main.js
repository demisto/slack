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
      <tr style={{height: '26px', padding: '30px'}}>
        <td style={{color: colors.date}}>{t}</td>
        <td style={{color: colors.channel, overflow: 'auto', padding: '15px'}}>#{msg.channel}</td>
        <td style={{color: colors.user, overflow: 'auto', padding: '15px'}}>@{msg.user}</td>
        <td style={{color: colors.text}}>{msg.text}</td>
      </tr>
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
      <tr style={{color: colors.dateSep, backgroundColor: colors.dateSepBack, textAlign: 'center', height: '26px', padding: '30px'}}>
        <td colSpan="4">{d.toDateString()}</td>
      </tr>
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
      <table role="presentation" style={{backgroundColor: colors.background, width: '100%', tableLayout: 'fixed'}}>
        <colgroup>
          <col style={{width: '8%'}} />
          <col style={{width: '15%'}} />
          <col style={{width: '15%'}} />
          <col />
        </colgroup>
        <tbody>
          {msgObjects}
        </tbody>
      </table>
    );
  }
}

ReactDOM.render(
  <div>
    <Frame />
  </div>,
  document.getElementById('app')
);
