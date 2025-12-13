class KindaVMClient {
    constructor() {
        this.ws = null;
        this.reconnectDelay = 1000;
        this.maxReconnectDelay = 30000;
        this.controlArea = document.getElementById('controlArea');
        this.mjpegFeed = document.getElementById('mjpegFeed');
        this.placeholder = document.getElementById('placeholder');
        this.isActive = false;
        this.pressedKeys = new Set();
        this.ustreamerPort = null; // Will be loaded from server config

        this.init();
    }

    init() {
        this.setupEventListeners();
        this.setupVideoControls();
        this.setupSystemControls();
        this.setupKeyboardShortcuts();
        this.setupSettings();
        this.loadHostname();
        this.loadConfig();
        this.loadSettings();
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

    async loadConfig() {
        try {
            const response = await fetch('/config');
            if (!response.ok) {
                console.warn('Failed to load config');
                return;
            }

            const data = await response.json();
            if (data.ustreamerPort) {
                this.ustreamerPort = data.ustreamerPort;
                console.log('ustreamer port:', this.ustreamerPort);
            }
        } catch (err) {
            console.error('Error loading config:', err);
        }
    }

    setupVideoControls() {
        const playControl = document.getElementById('playControl');
        const stopControl = document.getElementById('stopControl');
        const keyboardToggle = document.getElementById('keyboardToggle');
        const keyboardPopup = document.getElementById('keyboardPopup');
        const settingsPopup = document.getElementById('settingsPopup');

        playControl.addEventListener('click', (e) => {
            e.stopPropagation();
            keyboardPopup.classList.remove('show');
            settingsPopup.classList.remove('show');
            this.activateControl();
        });

        stopControl.addEventListener('click', (e) => {
            e.stopPropagation();
            keyboardPopup.classList.remove('show');
            settingsPopup.classList.remove('show');
            this.stopControl();
        });

        keyboardToggle.addEventListener('click', (e) => {
            e.stopPropagation();
            settingsPopup.classList.remove('show');
            keyboardPopup.classList.add('show');
        });

        keyboardPopup.addEventListener('click', (e) => {
            if (e.target === keyboardPopup) {
                keyboardPopup.classList.remove('show');
            }
        });
    }

    setupSettings() {
        const settingsToggle = document.getElementById('settingsToggle');
        const settingsPopup = document.getElementById('settingsPopup');
        const saveSettings = document.getElementById('saveSettings');
        const resetSettings = document.getElementById('resetSettings');
        const qualitySetting = document.getElementById('qualitySetting');
        const qualityValue = document.getElementById('qualityValue');
        const keyboardPopup = document.getElementById('keyboardPopup');

        this.originalSettings = null;

        const updateButtonStates = () => {
            if (!this.originalSettings) return;

            const currentSettings = {
                quality: parseInt(qualitySetting.value),
                desiredFps:
                    document.getElementById('fpsSetting').value === ''
                        ? 0
                        : parseInt(document.getElementById('fpsSetting').value),
            };

            const hasChanges =
                currentSettings.quality !== this.originalSettings.quality ||
                currentSettings.desiredFps !== this.originalSettings.desiredFps;

            saveSettings.disabled = !hasChanges;

            const isDefaults =
                parseInt(qualitySetting.value) === 80 && document.getElementById('fpsSetting').value === '';
            resetSettings.disabled = isDefaults;
        };

        settingsToggle.addEventListener('click', (e) => {
            e.stopPropagation();
            keyboardPopup.classList.remove('show');
            settingsPopup.classList.add('show');
        });

        settingsPopup.addEventListener('click', (e) => {
            if (e.target === settingsPopup) {
                settingsPopup.classList.remove('show');
            }
        });

        qualitySetting.addEventListener('input', (e) => {
            qualityValue.textContent = e.target.value;
            updateButtonStates();
        });

        document.getElementById('fpsSetting').addEventListener('input', updateButtonStates);

        saveSettings.addEventListener('click', (e) => {
            e.stopPropagation();
            this.saveSettings();
        });

        resetSettings.addEventListener('click', (e) => {
            e.stopPropagation();
            this.resetSettings();
        });
    }

    async loadSettings() {
        try {
            const response = await fetch('/settings');
            if (!response.ok) {
                console.warn('Failed to load settings');
                return;
            }

            const settings = await response.json();
            this.updateSettingsUI(settings);
        } catch (err) {
            console.error('Error loading settings:', err);
        }
    }

    updateSettingsUI(settings) {
        document.getElementById('qualitySetting').value = settings.quality;
        document.getElementById('qualityValue').textContent = settings.quality;
        document.getElementById('fpsSetting').value = settings.desiredFps || '';

        this.originalSettings = {
            quality: settings.quality,
            desiredFps: settings.desiredFps || 0,
        };

        const saveSettings = document.getElementById('saveSettings');
        const resetSettings = document.getElementById('resetSettings');

        saveSettings.disabled = true;

        const isDefaults = settings.quality === 80 && settings.desiredFps === 0;
        resetSettings.disabled = isDefaults;
    }

    async saveSettings() {
        const settings = {
            quality: parseInt(document.getElementById('qualitySetting').value),
            desiredFps: parseInt(document.getElementById('fpsSetting').value) || 0,
        };

        try {
            const response = await fetch('/settings/update', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify(settings),
            });

            if (!response.ok) {
                throw new Error('Failed to save settings');
            }

            console.log('Settings saved successfully');

            // Update original settings to current values
            this.updateSettingsUI(settings);

            document.getElementById('settingsPopup').classList.remove('show');

            // Restart video if it's currently running
            if (this.mjpegFeed.src) {
                await this.stopVideoStream();
                await this.startVideoStream();
            }
        } catch (err) {
            console.error('Error saving settings:', err);
            alert('Failed to save settings: ' + err.message);
        }
    }

    async resetSettings() {
        const defaultSettings = {
            quality: 80,
            desiredFps: 0,
        };

        this.updateSettingsUI(defaultSettings);

        try {
            const response = await fetch('/settings/update', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify(defaultSettings),
            });

            if (!response.ok) {
                throw new Error('Failed to reset settings');
            }

            console.log('Settings reset to defaults');
        } catch (err) {
            console.error('Error resetting settings:', err);
            alert('Failed to reset settings: ' + err.message);
        }
    }

    updateControlButtons() {
        const playControl = document.getElementById('playControl');
        const stopControl = document.getElementById('stopControl');

        if (this.isActive || this.mjpegFeed.src) {
            playControl.disabled = true;
            stopControl.disabled = false;
        } else {
            playControl.disabled = false;
            stopControl.disabled = true;
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
        const sendButton = document.getElementById('sendKeyCombo');
        const keyInput = document.getElementById('keyInput');
        const specialKeySelect = document.getElementById('specialKeySelect');

        const updateSendButtonState = () => {
            const hasKey = keyInput.value.trim() !== '';
            const hasSpecialKey = specialKeySelect.value !== '';
            sendButton.disabled = !hasKey && !hasSpecialKey;
        };

        // Clear the other input when one is used
        keyInput.addEventListener('input', () => {
            if (keyInput.value) {
                specialKeySelect.value = '';
            }
            updateSendButtonState();
        });

        specialKeySelect.addEventListener('change', () => {
            if (specialKeySelect.value) {
                keyInput.value = '';
            }
            updateSendButtonState();
        });

        sendButton.addEventListener('click', (e) => {
            e.stopPropagation();
            this.sendKeyCombo();
        });

        // Allow Enter to send the combo
        keyInput.addEventListener('keydown', (e) => {
            if (e.key === 'Enter') {
                e.preventDefault();
                this.sendKeyCombo();
            }
        });

        // Initialize button state
        updateSendButtonState();
    }

    sendKeyCombo() {
        const keyInput = document.getElementById('keyInput');
        const specialKeySelect = document.getElementById('specialKeySelect');
        const modCtrl = document.getElementById('modCtrl').checked;
        const modShift = document.getElementById('modShift').checked;
        const modAlt = document.getElementById('modAlt').checked;
        const modMeta = document.getElementById('modMeta').checked;

        // Determine which key to send
        let keyCode = '';

        if (keyInput.value) {
            // Regular key input (a-z, 0-9)
            const char = keyInput.value.toLowerCase();
            if (char >= 'a' && char <= 'z') {
                keyCode = 'Key' + char.toUpperCase();
            } else if (char >= '0' && char <= '9') {
                keyCode = 'Digit' + char;
            } else {
                console.warn('Invalid key input:', keyInput.value);
                return;
            }
        } else if (specialKeySelect.value) {
            // Special key from dropdown
            keyCode = specialKeySelect.value;
        } else {
            console.warn('No key selected');
            return;
        }

        // Build modifiers array
        const modifiers = [];
        if (modCtrl) modifiers.push('ctrl');
        if (modShift) modifiers.push('shift');
        if (modAlt) modifiers.push('alt');
        if (modMeta) modifiers.push('meta');

        // Send keydown and keyup events
        this.sendComboEvent('keydown', keyCode, modifiers);
        setTimeout(() => {
            this.sendComboEvent('keyup', keyCode, modifiers);
        }, 50);
    }

    sendComboEvent(eventType, keyCode, modifiers) {
        if (!this.ws || this.ws.readyState !== WebSocket.OPEN) {
            console.warn('WebSocket not connected, connecting now...');
            this.connect();

            const waitForConnection = () => {
                if (this.ws && this.ws.readyState === WebSocket.OPEN) {
                    this.sendEvent({
                        type: eventType,
                        code: keyCode,
                        modifiers: modifiers,
                    });
                } else {
                    setTimeout(waitForConnection, 100);
                }
            };
            setTimeout(waitForConnection, 100);
            return;
        }

        this.sendEvent({
            type: eventType,
            code: keyCode,
            modifiers: modifiers,
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
                this.updateControlButtons();
            } else {
                this.isActive = false;
                this.controlArea.classList.remove('active');
                this.releaseAllKeys();
                this.updateControlButtons();
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
        if (!this.ustreamerPort) {
            console.error('ustreamer port not configured');
            return;
        }

        try {
            // Start ustreamer on the server
            const response = await fetch('/video/start', { method: 'POST' });
            if (!response.ok) {
                throw new Error('Failed to start video stream');
            }

            // Give ustreamer a moment to start
            await new Promise((resolve) => setTimeout(resolve, 500));

            // Build URL using current hostname and configured port
            const ustreamerUrl = `http://${window.location.hostname}:${this.ustreamerPort}/stream`;
            this.mjpegFeed.src = ustreamerUrl;
            this.mjpegFeed.style.display = 'block';
            this.placeholder.style.display = 'none';
            this.updateControlButtons();
            console.log(`MJPEG streaming started from ustreamer: ${ustreamerUrl}`);
        } catch (err) {
            console.error('Failed to start MJPEG stream:', err);
        }
    }

    async stopVideoStream() {
        this.mjpegFeed.src = '';
        this.mjpegFeed.style.display = 'none';
        this.placeholder.style.display = 'flex';

        try {
            // Stop ustreamer on the server
            const response = await fetch('/video/stop', { method: 'POST' });
            if (!response.ok) {
                console.warn('Failed to stop video stream on server');
            }
            console.log('MJPEG streaming stopped');
        } catch (err) {
            console.error('Error stopping video stream:', err);
        }
    }

    stopControl() {
        // Exit pointer lock if active
        if (document.pointerLockElement === this.controlArea) {
            document.exitPointerLock();
        } else {
            // If pointer lock wasn't active, manually clear state
            this.isActive = false;
            this.controlArea.classList.remove('active');
            this.releaseAllKeys();
        }

        // Stop video and disconnect
        this.stopVideoStream();
        this.disconnect();

        // Update button states after everything is cleared
        this.updateControlButtons();
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
            modifiers: modifiers,
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
            modifiers: modifiers,
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
                y: e.movementY,
            });
        }
    }

    handleMouseDown(e) {
        if (!this.isActive) return;

        e.preventDefault();

        this.sendEvent({
            type: 'mousedown',
            button: this.getButtonName(e.button),
        });
    }

    handleMouseUp(e) {
        if (!this.isActive) return;

        e.preventDefault();

        this.sendEvent({
            type: 'mouseup',
            button: this.getButtonName(e.button),
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
                delta: delta,
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
            case 0:
                return 'left';
            case 1:
                return 'middle';
            case 2:
                return 'right';
            default:
                return 'left';
        }
    }

    releaseAllKeys() {
        // Send key up events for all pressed keys
        for (const code of this.pressedKeys) {
            this.sendEvent({
                type: 'keyup',
                code: code,
                modifiers: [],
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
