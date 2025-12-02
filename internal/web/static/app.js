class KindaVMClient {
    constructor() {
        this.ws = null;
        this.reconnectDelay = 1000;
        this.maxReconnectDelay = 30000;
        this.controlArea = document.getElementById('controlArea');
        this.statusIndicator = document.getElementById('statusIndicator');
        this.statusText = document.getElementById('statusText');
        this.isActive = false;
        this.pressedKeys = new Set();

        this.init();
    }

    init() {
        this.connect();
        this.setupEventListeners();
    }

    connect() {
        const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
        const wsUrl = `${protocol}//${window.location.host}/ws`;

        this.updateStatus('connecting', 'Connecting...');

        this.ws = new WebSocket(wsUrl);

        this.ws.onopen = () => {
            this.updateStatus('connected', 'Connected');
            this.reconnectDelay = 1000;
            console.log('WebSocket connected');
        };

        this.ws.onclose = () => {
            this.updateStatus('disconnected', 'Disconnected');
            console.log('WebSocket disconnected, reconnecting...');
            this.scheduleReconnect();
        };

        this.ws.onerror = (error) => {
            console.error('WebSocket error:', error);
            this.updateStatus('disconnected', 'Connection error');
        };
    }

    scheduleReconnect() {
        setTimeout(() => {
            this.connect();
            this.reconnectDelay = Math.min(this.reconnectDelay * 2, this.maxReconnectDelay);
        }, this.reconnectDelay);
    }

    updateStatus(state, text) {
        this.statusIndicator.className = `status-indicator ${state}`;
        this.statusText.textContent = text;
    }

    setupEventListeners() {
        // Click to activate control
        this.controlArea.addEventListener('click', () => {
            if (!this.isActive) {
                this.activateControl();
            }
        });

        // Handle pointer lock change
        document.addEventListener('pointerlockchange', () => {
            if (document.pointerLockElement === this.controlArea) {
                this.isActive = true;
                this.controlArea.classList.add('active');
            } else {
                this.isActive = false;
                this.controlArea.classList.remove('active');
                this.releaseAllKeys();
            }
        });

        // Warn before page unload (closing tab, reloading, etc.)
        window.addEventListener('beforeunload', (e) => {
            if (this.isActive || (this.ws && this.ws.readyState === WebSocket.OPEN)) {
                e.preventDefault();
                // Modern browsers ignore custom messages and show a generic one
                return '';
            }
        });

        // Keyboard events - capture on window level
        window.addEventListener('keydown', (e) => this.handleKeyDown(e));
        window.addEventListener('keyup', (e) => this.handleKeyUp(e));

        // Mouse events - only when active
        this.controlArea.addEventListener('mousemove', (e) => this.handleMouseMove(e));
        this.controlArea.addEventListener('mousedown', (e) => this.handleMouseDown(e));
        this.controlArea.addEventListener('mouseup', (e) => this.handleMouseUp(e));
        this.controlArea.addEventListener('wheel', (e) => this.handleWheel(e));
    }

    activateControl() {
        this.controlArea.requestPointerLock();
    }

    handleKeyDown(e) {
        if (!this.isActive) return;

        // Don't process if already pressed (key repeat)
        if (this.pressedKeys.has(e.code)) return;

        this.pressedKeys.add(e.code);

        // Prevent all browser shortcuts when control is active
        // This will block Ctrl-R (reload), Ctrl-S (save), Ctrl-T (new tab), etc.
        // Note: Ctrl-W (close tab) cannot be prevented for security reasons
        e.preventDefault();
        e.stopPropagation();

        const modifiers = this.getModifiers(e);

        this.sendEvent({
            type: 'keydown',
            code: e.code,
            modifiers: modifiers
        });
    }

    handleKeyUp(e) {
        if (!this.isActive) return;

        this.pressedKeys.delete(e.code);

        e.preventDefault();
        e.stopPropagation();

        const modifiers = this.getModifiers(e);

        this.sendEvent({
            type: 'keyup',
            code: e.code,
            modifiers: modifiers
        });
    }

    handleMouseMove(e) {
        if (!this.isActive) return;

        // movementX and movementY are the relative mouse movements
        // when pointer lock is active
        if (e.movementX !== 0 || e.movementY !== 0) {
            this.sendEvent({
                type: 'mousemove',
                x: e.movementX,
                y: e.movementY
            });
        }
    }

    handleMouseDown(e) {
        if (!this.isActive) return;

        e.preventDefault();

        this.sendEvent({
            type: 'mousedown',
            button: this.getButtonName(e.button)
        });
    }

    handleMouseUp(e) {
        if (!this.isActive) return;

        e.preventDefault();

        this.sendEvent({
            type: 'mouseup',
            button: this.getButtonName(e.button)
        });
    }

    handleWheel(e) {
        if (!this.isActive) return;

        e.preventDefault();

        // Normalize wheel delta (different browsers report differently)
        let delta = 0;
        if (e.deltaY < 0) {
            delta = -1; // Scroll up
        } else if (e.deltaY > 0) {
            delta = 1; // Scroll down
        }

        if (delta !== 0) {
            this.sendEvent({
                type: 'wheel',
                delta: delta
            });
        }
    }

    getModifiers(e) {
        const modifiers = [];
        if (e.ctrlKey) modifiers.push('ctrl');
        if (e.shiftKey) modifiers.push('shift');
        if (e.altKey) modifiers.push('alt');
        if (e.metaKey) modifiers.push('meta');
        return modifiers;
    }

    getButtonName(button) {
        switch (button) {
            case 0: return 'left';
            case 1: return 'middle';
            case 2: return 'right';
            default: return 'left';
        }
    }

    releaseAllKeys() {
        // Send key up events for all pressed keys
        for (const code of this.pressedKeys) {
            this.sendEvent({
                type: 'keyup',
                code: code,
                modifiers: []
            });
        }
        this.pressedKeys.clear();
    }

    sendEvent(event) {
        if (this.ws && this.ws.readyState === WebSocket.OPEN) {
            this.ws.send(JSON.stringify(event));
        } else {
            console.warn('WebSocket not connected, event not sent:', event);
        }
    }
}

// Initialize the client when the page loads
window.addEventListener('DOMContentLoaded', () => {
    new KindaVMClient();
});
