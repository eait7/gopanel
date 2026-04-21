/* ═══════════════════════════════════════════════════════════════
   GoPanel — Core Application Logic
   ═══════════════════════════════════════════════════════════════ */

const GoPanel = {
    currentSection: 'dashboard',
    refreshInterval: null,
    serviceLinks: {},

    // ── Initialization ──
    async init() {
        this.bindEvents();
        const session = await this.checkSession();
        if (session.authenticated) {
            this.showApp(session.username);
        } else {
            this.showLogin();
        }
    },

    // ── Event Binding ──
    bindEvents() {
        // Login form
        document.getElementById('login-form').addEventListener('submit', (e) => {
            e.preventDefault();
            this.login();
        });

        // Logout
        document.getElementById('logout-btn').addEventListener('click', () => this.logout());

        // Navigation
        document.querySelectorAll('.nav-item[data-section]').forEach(item => {
            item.addEventListener('click', (e) => {
                e.preventDefault();
                this.navigate(item.dataset.section);
            });
        });

        // Mobile menu
        const mobileBtn = document.getElementById('mobile-menu-btn');
        if (mobileBtn) {
            mobileBtn.addEventListener('click', () => this.toggleSidebar());
        }

        window.addEventListener('hashchange', () => {
            const hash = location.hash.slice(1) || 'dashboard';
            if (['dashboard', 'domains', 'containers', 'system', 'settings'].includes(hash)) {
                this.navigate(hash, false);
            }
        });

        // Close modals on overlay click
        document.querySelectorAll('.modal-overlay').forEach(overlay => {
            overlay.addEventListener('click', (e) => {
                if (e.target === overlay) {
                    overlay.style.display = 'none';
                }
            });
        });

        // Quick actions
        document.getElementById('action-open-files').addEventListener('click', () => {
            if (this.serviceLinks.filebrowser) {
                window.open(this.serviceLinks.filebrowser, '_blank');
            }
        });
        document.getElementById('action-open-portainer').addEventListener('click', () => {
            if (this.serviceLinks.portainer) {
                window.open(this.serviceLinks.portainer, '_blank');
            }
        });
    },

    // ── Auth ──
    async checkSession() {
        try {
            const resp = await fetch('/api/auth/session');
            return await resp.json();
        } catch {
            return { authenticated: false };
        }
    },

    async login() {
        const btn = document.getElementById('login-btn');
        const btnText = btn.querySelector('.btn-text');
        const btnLoader = btn.querySelector('.btn-loader');
        const errorEl = document.getElementById('login-error');

        btnText.style.display = 'none';
        btnLoader.style.display = 'inline-block';
        errorEl.style.display = 'none';

        try {
            const resp = await fetch('/api/auth/login', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    username: document.getElementById('login-username').value,
                    password: document.getElementById('login-password').value,
                }),
            });

            const data = await resp.json();
            if (resp.ok && data.success) {
                this.showApp(data.username);
            } else {
                errorEl.textContent = data.error || 'Invalid credentials';
                errorEl.style.display = 'block';
            }
        } catch (err) {
            errorEl.textContent = 'Connection error. Is the server running?';
            errorEl.style.display = 'block';
        } finally {
            btnText.style.display = 'inline';
            btnLoader.style.display = 'none';
        }
    },

    async logout() {
        await fetch('/api/auth/logout', { method: 'POST' });
        this.stopRefresh();
        this.showLogin();
    },

    // ── Screen Switching ──
    showLogin() {
        document.getElementById('login-screen').style.display = 'flex';
        document.getElementById('app').style.display = 'none';
    },

    showApp(username) {
        document.getElementById('login-screen').style.display = 'none';
        document.getElementById('app').style.display = 'flex';

        // Set user info
        document.getElementById('user-name').textContent = username;
        document.getElementById('user-avatar').textContent = username.charAt(0).toUpperCase();

        // Load data
        this.loadServiceLinks();
        this.navigate(location.hash.slice(1) || 'dashboard', false);
        this.startRefresh();
    },

    // ── Navigation ──
    navigate(section, updateHash = true) {
        if (updateHash) {
            location.hash = section;
        }

        this.currentSection = section;

        // Update nav items
        document.querySelectorAll('.nav-item[data-section]').forEach(item => {
            item.classList.toggle('active', item.dataset.section === section);
        });

        // Show section
        document.querySelectorAll('.section').forEach(sec => {
            sec.classList.remove('active');
        });
        const target = document.getElementById(`section-${section}`);
        if (target) {
            target.classList.add('active');
        }

        // Load section data
        this.loadSectionData(section);

        // Close mobile sidebar
        document.querySelector('.sidebar').classList.remove('open');
    },

    loadSectionData(section) {
        switch (section) {
            case 'dashboard':
                SystemModule.loadStats();
                ContainersModule.loadContainers();
                DomainsModule.loadDomains();
                break;
            case 'domains':
                DomainsModule.loadDomains();
                break;
            case 'containers':
                ContainersModule.loadContainers();
                break;
            case 'system':
                SystemModule.loadStats();
                break;
        }
    },

    // ── Service Links ──
    async loadServiceLinks() {
        try {
            const resp = await fetch('/api/links');
            if (resp.ok) {
                this.serviceLinks = await resp.json();

                // Update nav links
                const fbLink = document.getElementById('nav-filebrowser');
                const ptLink = document.getElementById('nav-portainer');
                if (this.serviceLinks.filebrowser) {
                    fbLink.href = this.serviceLinks.filebrowser;
                }
                if (this.serviceLinks.portainer) {
                    ptLink.href = this.serviceLinks.portainer;
                }
            }
        } catch {
            // Silent fail — links stay as #
        }
    },

    // ── Auto-Refresh ──
    startRefresh() {
        this.stopRefresh();
        this.refreshInterval = setInterval(() => {
            if (this.currentSection === 'dashboard' || this.currentSection === 'system') {
                SystemModule.loadStats();
            }
        }, 5000);
    },

    stopRefresh() {
        if (this.refreshInterval) {
            clearInterval(this.refreshInterval);
            this.refreshInterval = null;
        }
    },

    // ── Sidebar Toggle (Mobile) ──
    toggleSidebar() {
        document.querySelector('.sidebar').classList.toggle('open');
    },

    // ── Toast Notifications ──
    toast(message, type = 'info') {
        const container = document.getElementById('toast-container');
        const toast = document.createElement('div');
        toast.className = `toast toast-${type}`;
        toast.innerHTML = `
            <span>${this.escapeHtml(message)}</span>
        `;
        container.appendChild(toast);

        setTimeout(() => {
            toast.classList.add('removing');
            setTimeout(() => toast.remove(), 300);
        }, 4000);
    },

    // ── Confirm Dialog ──
    confirm(title, message) {
        return new Promise((resolve) => {
            document.getElementById('confirm-title').textContent = title;
            document.getElementById('confirm-message').textContent = message;
            document.getElementById('confirm-modal').style.display = 'flex';

            const ok = document.getElementById('confirm-ok');
            const cancel = document.getElementById('confirm-cancel');
            const close = document.getElementById('confirm-close');

            const cleanup = () => {
                document.getElementById('confirm-modal').style.display = 'none';
                ok.replaceWith(ok.cloneNode(true));
                cancel.replaceWith(cancel.cloneNode(true));
                close.replaceWith(close.cloneNode(true));
            };

            ok.addEventListener('click', () => { cleanup(); resolve(true); }, { once: true });
            cancel.addEventListener('click', () => { cleanup(); resolve(false); }, { once: true });
            close.addEventListener('click', () => { cleanup(); resolve(false); }, { once: true });
        });
    },

    // ── Utility ──
    escapeHtml(str) {
        const div = document.createElement('div');
        div.textContent = str;
        return div.innerHTML;
    },

    formatBytes(bytes) {
        if (bytes === 0) return '0 B';
        const units = ['B', 'KB', 'MB', 'GB', 'TB'];
        const i = Math.floor(Math.log(bytes) / Math.log(1024));
        return (bytes / Math.pow(1024, i)).toFixed(1) + ' ' + units[i];
    },

    async api(url, options = {}) {
        return this.apiRequest(url, options);
    },

    async apiRequest(url, options = {}) {
        try {
            const resp = await fetch(url, {
                ...options,
                headers: {
                    'Content-Type': 'application/json',
                    ...options.headers,
                },
            });
            const data = await resp.json();
            if (!resp.ok) {
                throw new Error(data.error || `HTTP ${resp.status}`);
            }
            return data;
        } catch (err) {
            if (err.message.includes('unauthorized') || err.message.includes('session expired')) {
                this.showLogin();
                return null;
            }
            throw err;
        }
    }
};

// ── Boot ──
document.addEventListener('DOMContentLoaded', () => GoPanel.init());
