/* ═══════════════════════════════════════════════════════════════
   GoPanel — Domain Management Module
   ═══════════════════════════════════════════════════════════════ */

const DomainsModule = {
    domains: [],

    async loadDomains() {
        try {
            const data = await GoPanel.apiRequest('/api/domains');
            if (!data) return;

            this.domains = data.domains || [];
            this.renderDomains();
            this.updateDashboardStat();
        } catch (err) {
            console.error('Failed to load domains:', err);
        }
    },

    renderDomains() {
        const tbody = document.getElementById('domains-tbody');
        const empty = document.getElementById('domains-empty');
        const tableWrap = document.getElementById('domains-table-wrap');

        if (!this.domains.length) {
            empty.style.display = 'block';
            tableWrap.style.display = 'none';
            return;
        }

        empty.style.display = 'none';
        tableWrap.style.display = 'block';

        tbody.innerHTML = this.domains.map((d, i) => `
            <tr>
                <td>
                    <span class="domain-name">${GoPanel.escapeHtml(d.domains ? d.domains.join(', ') : '—')}</span>
                </td>
                <td>
                    <code style="font-family:var(--font-mono);font-size:13px;color:var(--text-secondary)">${GoPanel.escapeHtml(d.upstream || '—')}</code>
                </td>
                <td>
                    <span class="badge ${d.type === 'reverse_proxy' ? 'badge-blue' : 'badge-amber'}">
                        ${d.type === 'reverse_proxy' ? '⇄ Proxy' : '📁 Files'}
                    </span>
                </td>
                <td>
                    <span class="ssl-active">
                        <span class="ssl-dot"></span>Auto SSL
                    </span>
                </td>
                <td>
                    <div class="table-actions">
                        <button class="btn-icon" title="Edit" onclick="DomainsModule.editDomain(${d.id})">
                            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"/><path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"/></svg>
                        </button>
                        <button class="btn-icon" title="Delete" onclick="DomainsModule.deleteDomain(${d.id}, '${GoPanel.escapeHtml(d.domains ? d.domains[0] : '')}')">
                            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="var(--red)" stroke-width="2"><polyline points="3 6 5 6 21 6"/><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/></svg>
                        </button>
                    </div>
                </td>
            </tr>
        `).join('');
    },

    updateDashboardStat() {
        const el = document.getElementById('stat-domains');
        if (el) el.textContent = this.domains.length;
    },

    // ── Add Domain Modal ──
    init() {
        document.getElementById('add-domain-btn').addEventListener('click', () => this.openModal());
        document.getElementById('domain-modal-close').addEventListener('click', () => this.closeModal());
        document.getElementById('domain-cancel-btn').addEventListener('click', () => this.closeModal());
        document.getElementById('domain-form').addEventListener('submit', (e) => {
            e.preventDefault();
            this.saveDomain();
        });

        // Update label on type change
        document.getElementById('domain-type').addEventListener('change', (e) => {
            const label = document.getElementById('domain-upstream-label');
            const hint = document.getElementById('domain-upstream-hint');
            const input = document.getElementById('domain-upstream');
            if (e.target.value === 'file_server') {
                label.textContent = 'Root Directory';
                hint.textContent = 'The directory to serve files from';
                input.placeholder = '/var/www/html';
            } else {
                label.textContent = 'Upstream (host:port)';
                hint.textContent = 'The backend server to proxy requests to';
                input.placeholder = 'localhost:8080';
            }
        });
    },

    openModal(editId = null) {
        const modal = document.getElementById('domain-modal');
        const title = document.getElementById('domain-modal-title');
        const saveBtn = document.getElementById('domain-save-btn');
        const editField = document.getElementById('domain-edit-id');

        if (editId !== null) {
            const domain = this.domains.find(d => d.id === editId);
            if (domain) {
                title.textContent = 'Edit Domain';
                saveBtn.textContent = 'Save Changes';
                document.getElementById('domain-name').value = domain.domains ? domain.domains[0] : '';
                document.getElementById('domain-upstream').value = domain.upstream || '';
                document.getElementById('domain-type').value = domain.type || 'reverse_proxy';
                editField.value = editId;
            }
        } else {
            title.textContent = 'Add Domain';
            saveBtn.textContent = 'Add Domain';
            document.getElementById('domain-form').reset();
            editField.value = '';
        }

        modal.style.display = 'flex';
    },

    closeModal() {
        document.getElementById('domain-modal').style.display = 'none';
    },

    async saveDomain() {
        const editId = document.getElementById('domain-edit-id').value;
        const payload = {
            domain: document.getElementById('domain-name').value.trim(),
            upstream: document.getElementById('domain-upstream').value.trim(),
            type: document.getElementById('domain-type').value,
        };

        if (!payload.domain || !payload.upstream) {
            GoPanel.toast('Please fill in all fields', 'error');
            return;
        }

        try {
            if (editId) {
                await GoPanel.apiRequest(`/api/domains/${editId}`, {
                    method: 'PUT',
                    body: JSON.stringify(payload),
                });
                GoPanel.toast('Domain updated successfully', 'success');
            } else {
                await GoPanel.apiRequest('/api/domains', {
                    method: 'POST',
                    body: JSON.stringify(payload),
                });
                GoPanel.toast('Domain added successfully! SSL will be provisioned automatically.', 'success');
            }

            this.closeModal();
            await this.loadDomains();
        } catch (err) {
            GoPanel.toast(err.message, 'error');
        }
    },

    editDomain(id) {
        this.openModal(id);
    },

    async deleteDomain(id, name) {
        const confirmed = await GoPanel.confirm(
            'Delete Domain',
            `Are you sure you want to remove "${name}"? The SSL certificate will also be removed.`
        );
        if (!confirmed) return;

        try {
            await GoPanel.apiRequest(`/api/domains/${id}`, { method: 'DELETE' });
            GoPanel.toast('Domain removed successfully', 'success');
            await this.loadDomains();
        } catch (err) {
            GoPanel.toast(err.message, 'error');
        }
    },
};

// Initialize when DOM is ready
document.addEventListener('DOMContentLoaded', () => DomainsModule.init());
