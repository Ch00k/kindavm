class H264Player {
  constructor(videoElement, wsUrl) {
    this.video = videoElement;
    this.wsUrl = wsUrl;
    this.ws = null;
    this.jmuxer = null;
    this.reconnectDelay = 1000;
    this.maxReconnectDelay = 30000;
    this.shouldReconnect = true;
  }

  async start() {
    console.log("Starting H264 player with jMuxer");

    this.shouldReconnect = true;

    this.jmuxer = new JMuxer({
      node: this.video,
      mode: "video",
      flushingTime: 0,
      fps: 30,
      debug: false,
      onReady: () => {
        console.log("jMuxer ready");
      },
    });

    this.connect();
    return true;
  }

  connect() {
    console.log("Connecting to H264 stream...");
    this.ws = new WebSocket(this.wsUrl);
    this.ws.binaryType = "arraybuffer";

    this.ws.onopen = () => {
      console.log("H264 WebSocket connected");
      this.reconnectDelay = 1000;
    };

    this.ws.onmessage = (event) => {
      if (this.jmuxer && event.data instanceof ArrayBuffer) {
        this.jmuxer.feed({
          video: new Uint8Array(event.data),
        });
      }
    };

    this.ws.onclose = () => {
      console.log("H264 WebSocket disconnected");
      if (this.shouldReconnect) {
        console.log("Reconnecting...");
        setTimeout(() => {
          this.connect();
          this.reconnectDelay = Math.min(
            this.reconnectDelay * 2,
            this.maxReconnectDelay,
          );
        }, this.reconnectDelay);
      }
    };

    this.ws.onerror = (error) => {
      console.error("H264 WebSocket error:", error);
    };
  }

  stop() {
    this.shouldReconnect = false;

    if (this.ws) {
      this.ws.close();
      this.ws = null;
    }

    if (this.jmuxer) {
      this.jmuxer.destroy();
      this.jmuxer = null;
    }
  }
}
