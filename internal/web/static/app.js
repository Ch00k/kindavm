class KindaVMClient {
    constructor() {
        this.ws = null;
        this.reconnectDelay = 1000;
        this.maxReconnectDelay = 30000;
        this.controlArea = document.getElementById('controlArea');
        this.videoFeed = document.getElementById('videoFeed');
        this.placeholder = document.getElementById('placeholder');
        this.isActive = false;
        this.pressedKeys = new Set();
        this.h264Player = null;
        this.videoSettings = {
            width: 0,
            height: 0,
            framerate: 30
        };

        this.init();
    }

    init() {
        this.setupEventListeners();
        this.loadCameraModes();
        this.setupVideoControls();
        this.setupSystemControls();
        this.setupKeyboardShortcuts();
        this.loadHostname();
    }

    async loadHostname() {
        try {
            const response = await fetch('/hostname');
            if (!response.ok) {
                console.warn('Failed to load hostname');
                return;
            }

            const data = await response.json();
            if (data.hostname) {
                document.getElementById('hostname').textContent = data.hostname;
            }
        } catch (err) {
            console.error('Error loading hostname:', err);
        }
    }

    async loadCameraModes() {
        try {
            const response = await fetch('/camera-modes');
            if (!response.ok) {
                console.warn('Failed to load camera modes, using defaults');
                return;
            }

            const modes = await response.json();
            if (!modes || modes.length === 0) {
                console.warn('No camera modes returned, using defaults');
                return;
            }

            this.populateResolutionDropdown(modes);
        } catch (err) {
            console.error('Error loading camera modes:', err);
        }
    }

    populateResolutionDropdown(modes) {
        const resolutionSelect = document.getElementById('resolution');

        // Clear existing options except default
        resolutionSelect.innerHTML = '<option value="default" selected>Default (camera native)</option>';

        // Add detected modes
        modes.forEach(mode => {
            const option = document.createElement('option');
            option.value = `${mode.Width}x${mode.Height}`;
            option.textContent = `${mode.Width}x${mode.Height}`;
            resolutionSelect.appendChild(option);
        });

        console.log(`Loaded ${modes.length} camera modes`);
    }

    setupVideoControls() {
        const resolutionSelect = document.getElementById('resolution');
        const framerateSelect = document.getElementById('framerate');
        const expandToggle = document.getElementById('expandToggle');
        const closeExpand = document.getElementById('closeExpand');

        // Update settings when dropdowns change
        const updateSettings = () => {
            const resolution = resolutionSelect.value;
            if (resolution === 'default') {
                this.videoSettings.width = 0;
                this.videoSettings.height = 0;
            } else {
                const parts = resolution.split('x');
                this.videoSettings.width = parseInt(parts[0]);
                this.videoSettings.height = parseInt(parts[1]);
            }
            this.videoSettings.framerate = parseInt(framerateSelect.value);
        };

        resolutionSelect.addEventListener('change', updateSettings);
        framerateSelect.addEventListener('change', updateSettings);

        expandToggle.addEventListener('click', (e) => {
            e.stopPropagation();
            this.toggleExpanded();
        });

        closeExpand.addEventListener('click', (e) => {
            e.stopPropagation();
            this.toggleExpanded();
        });
    }

    toggleExpanded() {
        const isExpanded = this.controlArea.classList.contains('expanded');
        const closeButton = document.getElementById('closeExpand');
        const expandButton = document.getElementById('expandToggle');

        if (isExpanded) {
            this.controlArea.classList.remove('expanded');
            expandButton.classList.remove('expanded');
            expandButton.title = 'Expand video to viewport';
            expandButton.innerHTML = '<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M8 3H5a2 2 0 0 0-2 2v3"/><path d="M21 8V5a2 2 0 0 0-2-2h-3"/><path d="M3 16v3a2 2 0 0 0 2 2h3"/><path d="M16 21h3a2 2 0 0 0 2-2v-3"/></svg>';
            closeButton.style.display = 'none';
        } else {
            this.controlArea.classList.add('expanded');
            expandButton.classList.add('expanded');
            expandButton.title = 'Restore video size';
            expandButton.innerHTML = '<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M8 3v3a2 2 0 0 1-2 2H3"/><path d="M21 8h-3a2 2 0 0 1-2-2V3"/><path d="M3 16h3a2 2 0 0 1 2 2v3"/><path d="M16 21v-3a2 2 0 0 1 2-2h3"/></svg>';
            closeButton.style.display = 'flex';
        }
    }

    setupSystemControls() {
        const brightnessUpBtn = document.getElementById('brightnessUp');
        const brightnessDownBtn = document.getElementById('brightnessDown');
        const volumeUpBtn = document.getElementById('volumeUp');
        const volumeDownBtn = document.getElementById('volumeDown');

        brightnessUpBtn.addEventListener('click', (e) => {
            e.stopPropagation();
            this.sendSystemEvent('brightness_up');
        });

        brightnessDownBtn.addEventListener('click', (e) => {
            e.stopPropagation();
            this.sendSystemEvent('brightness_down');
        });

        volumeUpBtn.addEventListener('click', (e) => {
            e.stopPropagation();
            this.sendSystemEvent('volume_up');
        });

        volumeDownBtn.addEventListener('click', (e) => {
            e.stopPropagation();
            this.sendSystemEvent('volume_down');
        });
    }

    sendSystemEvent(eventType) {
        if (!this.ws || this.ws.readyState !== WebSocket.OPEN) {
            console.warn('WebSocket not connected, connecting now...');
            this.connect();

            // Wait for connection and retry
            const waitForConnection = () => {
                if (this.ws && this.ws.readyState === WebSocket.OPEN) {
                    this.sendEvent({ type: eventType });
                } else {
                    setTimeout(waitForConnection, 100);
                }
            };
            setTimeout(waitForConnection, 100);
            return;
        }

        this.sendEvent({ type: eventType });
    }

    setupKeyboardShortcuts() {
        const shortcuts = [
            { id: 'ctrlW', event: 'ctrl_w' },
            { id: 'ctrlT', event: 'ctrl_t' },
            { id: 'ctrlN', event: 'ctrl_n' },
            { id: 'ctrlQ', event: 'ctrl_q' },
            { id: 'ctrlTab', event: 'ctrl_tab' },
            { id: 'ctrlShiftTab', event: 'ctrl_shift_tab' },
            { id: 'ctrlShiftT', event: 'ctrl_shift_t' },
            { id: 'ctrlF4', event: 'ctrl_f4' },
            { id: 'altF4', event: 'alt_f4' },
            { id: 'f11', event: 'f11' }
        ];

        shortcuts.forEach(shortcut => {
            const button = document.getElementById(shortcut.id);
            if (button) {
                button.addEventListener('click', (e) => {
                    e.stopPropagation();
                    this.sendSystemEvent(shortcut.event);
                });
            }
        });
    }

    connect() {
        if (this.ws) {
            return;
        }

        const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
        const wsUrl = `${protocol}//${window.location.host}/ws`;

        this.ws = new WebSocket(wsUrl);

        this.ws.onopen = () => {
            this.reconnectDelay = 1000;
            console.log('Control WebSocket connected');
        };

        this.ws.onclose = () => {
            this.ws = null;
            console.log('Control WebSocket disconnected');
        };

        this.ws.onerror = (error) => {
            console.error('Control WebSocket error:', error);
        };
    }

    disconnect() {
        if (this.ws) {
            this.ws.close();
            this.ws = null;
        }
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
                this.connect();
            } else {
                this.isActive = false;
                this.controlArea.classList.remove('active');
                this.releaseAllKeys();
                this.stopVideoStream();
                this.disconnect();

                // Exit expanded mode when ESC is pressed
                if (this.controlArea.classList.contains('expanded')) {
                    this.toggleExpanded();
                }
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
        this.startVideoStream();
    }

    async startVideoStream() {
        if (this.h264Player) {
            return;
        }

        const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
        const params = new URLSearchParams({
            width: this.videoSettings.width,
            height: this.videoSettings.height,
            framerate: this.videoSettings.framerate
        });
        const h264Url = `${protocol}//${window.location.host}/video-stream?${params}`;

        try {
            this.h264Player = new H264Player(this.videoFeed, h264Url);
            const started = await this.h264Player.start();

            if (started) {
                console.log(`H264 streaming started (${this.videoSettings.width}x${this.videoSettings.height} @ ${this.videoSettings.framerate}fps)`);
                this.videoFeed.style.display = 'block';
                this.placeholder.style.display = 'none';

                this.videoFeed.addEventListener('loadedmetadata', () => {
                    console.log('H264 video metadata loaded');
                });

                this.videoFeed.addEventListener('error', (e) => {
                    console.error('H264 video error:', e);
                });
            }
        } catch (err) {
            console.error('Failed to start H264 stream:', err);
        }
    }

    stopVideoStream() {
        if (this.h264Player) {
            this.h264Player.stop();
            this.h264Player = null;
            this.videoFeed.style.display = 'none';
            this.placeholder.style.display = 'flex';
            console.log('H264 streaming stopped');
        }
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
window.addEventListener("DOMContentLoaded", () => {
    new KindaVMClient();
});
