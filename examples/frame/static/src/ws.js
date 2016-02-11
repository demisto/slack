let ws;

/**
 * Get server url based on HTTP / HTTPS
 * @returns {String} the WS URL
 */
function getWSUrl() {
  let proto;
  if (location.protocol === 'https:') {
    proto = 'wss://';
  } else {
    proto = 'ws://';
  }
  return proto + location.host + '/ws';
}

/**
 * Open web socket and register for events
 * @param {Function} handler - the handler for the message
 * @returns {Void} nothing
 */
export function open(handler) {
  if (window.WebSocket) {
    if (ws) {
      return;
    }
    ws = new WebSocket(getWSUrl());
    ws.onclose = () => {
      ws = null;
      setTimeout(open.bind(this, handler), 3000);
    };
    ws.onmessage = (message) => {
      var msg = JSON.parse(message.data);
      if (msg) {
        handler(msg);
      }
    };
  }
}

/**
 * Disconnect from web socket
 * @returns {Void} nothing
 */
export function disconnect() {
  if (ws) {
    ws.close();
    ws = null;
  }
}
