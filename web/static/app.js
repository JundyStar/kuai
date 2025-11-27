let currentTemplate = null;
let modal = null;
let allTemplates = []; // 存储所有模板用于搜索过滤

// Tab switching
document.querySelectorAll('.nav-tab').forEach(tab => {
    tab.addEventListener('click', (e) => {
        const tabName = tab.dataset.tab;
        showTab(tabName);
    });
});

function showTab(tab) {
    // Update tabs
    document.querySelectorAll('.nav-tab').forEach(t => t.classList.remove('active'));
    const activeTab = document.querySelector(`[data-tab="${tab}"]`);
    if (activeTab) {
        activeTab.classList.add('active');
    }
    
    // Update panels
    document.querySelectorAll('.panel').forEach(p => p.classList.remove('active'));
    const activePanel = document.getElementById(`panel-${tab}`);
    if (activePanel) {
        activePanel.classList.add('active');
    }
    
    // Scroll to main content
    if (activePanel) {
        activePanel.scrollIntoView({ behavior: 'smooth', block: 'start' });
    }
    
    // Load templates when switching to templates tab
    if (tab === 'templates') {
        loadTemplates();
    }
}

// Load templates
async function loadTemplates() {
    const container = document.getElementById('templates-grid');
    // 显示骨架屏
    container.innerHTML = `
        <div class="skeleton-card">
            <div class="skeleton-title"></div>
            <div class="skeleton-text"></div>
            <div class="skeleton-text"></div>
            <div class="skeleton-button"></div>
        </div>
        <div class="skeleton-card">
            <div class="skeleton-title"></div>
            <div class="skeleton-text"></div>
            <div class="skeleton-text"></div>
            <div class="skeleton-button"></div>
        </div>
        <div class="skeleton-card">
            <div class="skeleton-title"></div>
            <div class="skeleton-text"></div>
            <div class="skeleton-text"></div>
            <div class="skeleton-button"></div>
        </div>
    `;
    
    try {
        const res = await fetch('/api/templates');
        if (!res.ok) throw new Error('Failed to load templates');
        
        const templates = await res.json();
        allTemplates = templates; // 保存所有模板
        renderTemplates(templates);
    } catch (error) {
        container.innerHTML = `
            <div class="empty-state">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" style="color: var(--error);">
                    <circle cx="12" cy="12" r="10"/>
                    <line x1="12" y1="8" x2="12" y2="12"/>
                    <line x1="12" y1="16" x2="12.01" y2="16"/>
                </svg>
                <p style="color: var(--error);">加载失败</p>
                <p class="empty-hint">${escapeHtml(error.message)}</p>
            </div>
        `;
    }
}

// Open generate modal
async function openGenerateModal(templateName) {
    currentTemplate = templateName;
    document.getElementById('modal-title').textContent = `生成项目 - ${escapeHtml(templateName)}`;
    
    const messageDiv = document.getElementById('generate-message');
    messageDiv.innerHTML = '';
    
    try {
        const res = await fetch(`/api/templates/${encodeURIComponent(templateName)}`);
        if (!res.ok) throw new Error('Failed to load template details');
        
        const manifest = await res.json();
        const formFields = document.getElementById('form-fields');
        formFields.innerHTML = '';
        
        if (manifest.fields && manifest.fields.length > 0) {
            manifest.fields.forEach(field => {
                const div = document.createElement('div');
                div.className = 'form-group';
                div.innerHTML = `
                    <label class="form-label">
                        ${escapeHtml(field.prompt || field.name)}
                        ${field.description ? `<span style="color: var(--text-secondary); font-weight: normal;">(${escapeHtml(field.description)})</span>` : ''}
                        ${field.required ? '<span style="color: var(--error);">*</span>' : ''}
                    </label>
                    <input 
                        type="text" 
                        name="${escapeHtml(field.name)}" 
                        class="form-input"
                        value="${escapeHtml(field.default || '')}"
                        ${field.required ? 'required' : ''}
                        placeholder="${field.default ? '默认: ' + escapeHtml(field.default) : ''}"
                    >
                `;
                formFields.appendChild(div);
            });
        } else {
            formFields.innerHTML = '<p style="color: var(--text-secondary);">此模板无需配置参数</p>';
        }
        
        document.getElementById('generate-modal').classList.add('active');
    } catch (error) {
        alert('加载模板详情失败: ' + error.message);
    }
}

// Close modal
function closeModal() {
    document.getElementById('generate-modal').classList.remove('active');
    currentTemplate = null;
}

// Submit generate
async function submitGenerate() {
    const form = document.getElementById('generate-form');
    const formData = new FormData(form);
    const values = {};
    formData.forEach((value, key) => {
        values[key] = value;
    });
    
    const messageDiv = document.getElementById('generate-message');
    const submitBtn = document.querySelector('.modal-footer .btn-primary');
    const originalBtnText = submitBtn.innerHTML;
    
    // 显示加载状态
    submitBtn.classList.add('loading');
    submitBtn.disabled = true;
    messageDiv.innerHTML = '<div class="alert alert-info">正在生成项目...</div>';
    
    try {
        const res = await fetch('/api/generate', {
            method: 'POST',
            headers: {'Content-Type': 'application/json'},
            body: JSON.stringify({
                templateName: currentTemplate,
                values: values
            })
        });
        
        const result = await res.json();
        if (result.status === 'success') {
            messageDiv.innerHTML = '<div class="alert alert-success">✅ 项目生成成功！正在下载...</div>';
            showToast('项目生成成功，正在下载...', 'success');
            window.location.href = result.downloadUrl;
            setTimeout(() => {
                closeModal();
                messageDiv.innerHTML = '';
            }, 2000);
        } else {
            messageDiv.innerHTML = `<div class="alert alert-error">生成失败: ${escapeHtml(result.error || '未知错误')}</div>`;
            submitBtn.classList.remove('loading');
            submitBtn.disabled = false;
        }
    } catch (error) {
        messageDiv.innerHTML = `<div class="alert alert-error">生成失败: ${escapeHtml(error.message)}</div>`;
        submitBtn.classList.remove('loading');
        submitBtn.disabled = false;
    }
}

// Upload form
document.getElementById('upload-form').addEventListener('submit', async (e) => {
    e.preventDefault();
    const formData = new FormData(e.target);
    
    // 确保 force checkbox 的值正确设置
    const forceCheckbox = document.getElementById('force');
    if (forceCheckbox.checked) {
        formData.set('force', 'true');
    } else {
        formData.delete('force');
    }
    
    const messageDiv = document.getElementById('upload-message');
    const submitBtn = e.target.querySelector('button[type="submit"]');
    const originalBtnText = submitBtn.innerHTML;
    
    // 显示加载状态
    submitBtn.classList.add('loading');
    submitBtn.disabled = true;
    messageDiv.innerHTML = '<div class="alert alert-info">正在上传...</div>';
    
    try {
        const res = await fetch('/api/upload', {
            method: 'POST',
            body: formData
        });
        
        const result = await res.json();
        if (res.ok && result.status === 'success') {
            messageDiv.innerHTML = '<div class="alert alert-success">✅ ' + escapeHtml(result.message) + '</div>';
            e.target.reset();
            document.getElementById('file-name').textContent = '未选择文件';
            submitBtn.classList.remove('loading');
            submitBtn.disabled = false;
            // 显示成功提示
            showToast('模板上传成功！', 'success');
            loadTemplates();
            showTab('templates');
        } else {
            messageDiv.innerHTML = `<div class="alert alert-error">上传失败: ${escapeHtml(result.error || '未知错误')}</div>`;
            submitBtn.classList.remove('loading');
            submitBtn.disabled = false;
        }
    } catch (error) {
        messageDiv.innerHTML = `<div class="alert alert-error">上传失败: ${escapeHtml(error.message)}</div>`;
        submitBtn.classList.remove('loading');
        submitBtn.disabled = false;
    }
});

// File input
document.getElementById('file-input').addEventListener('change', (e) => {
    const fileName = e.target.files[0]?.name || '未选择文件';
    document.getElementById('file-name').textContent = fileName;
});

// Drag and drop upload
const fileUploadArea = document.getElementById('file-upload-area');
const fileInput = document.getElementById('file-input');
const fileUploadLabel = document.getElementById('file-upload-label');

['dragenter', 'dragover', 'dragleave', 'drop'].forEach(eventName => {
    fileUploadArea.addEventListener(eventName, preventDefaults, false);
});

function preventDefaults(e) {
    e.preventDefault();
    e.stopPropagation();
}

['dragenter', 'dragover'].forEach(eventName => {
    fileUploadArea.addEventListener(eventName, () => {
        fileUploadArea.classList.add('drag-over');
    }, false);
});

['dragleave', 'drop'].forEach(eventName => {
    fileUploadArea.addEventListener(eventName, () => {
        fileUploadArea.classList.remove('drag-over');
    }, false);
});

fileUploadArea.addEventListener('drop', (e) => {
    const dt = e.dataTransfer;
    const files = dt.files;
    
    if (files.length > 0) {
        const file = files[0];
        if (file.name.endsWith('.zip')) {
            fileInput.files = files;
            document.getElementById('file-name').textContent = file.name;
            showToast('文件已选择', 'success');
        } else {
            showToast('请选择 ZIP 格式的文件', 'error');
        }
    }
}, false);

// Keyboard shortcuts
document.addEventListener('keydown', (e) => {
    // ESC to close modals
    if (e.key === 'Escape') {
        const generateModal = document.getElementById('generate-modal');
        const previewModal = document.getElementById('preview-modal');
        if (generateModal.classList.contains('active')) {
            closeModal();
        } else if (previewModal.classList.contains('active')) {
            closePreviewModal();
        }
    }
    
    // Enter to submit forms (when in modal)
    if (e.key === 'Enter' && e.ctrlKey) {
        const generateModal = document.getElementById('generate-modal');
        if (generateModal.classList.contains('active')) {
            e.preventDefault();
            submitGenerate();
        }
    }
});

// Render templates
function renderTemplates(templates) {
    const container = document.getElementById('templates-grid');
    container.innerHTML = '';
    
    if (templates.length === 0) {
        container.innerHTML = `
            <div class="empty-state">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
                    <path d="M12 2L2 7l10 5 10-5-10-5z"/>
                    <path d="M2 17l10 5 10-5M2 12l10 5 10-5"/>
                </svg>
                <p>暂无模板</p>
                <p class="empty-hint">点击上方"上传"标签添加你的第一个模板</p>
            </div>
        `;
        return;
    }
    
    templates.forEach(tpl => {
        const card = document.createElement('div');
        card.className = 'template-card';
        card.setAttribute('data-template-name', tpl.Name);
        card.innerHTML = `
            <div class="template-card-header">
                <h3>${escapeHtml(tpl.Name)}</h3>
                <button class="template-preview-btn" onclick="openPreviewModal('${escapeHtml(tpl.Name)}')" title="查看详情">
                    <svg width="18" height="18" viewBox="0 0 18 18" fill="none" stroke="currentColor">
                        <path d="M1 9s2-4 8-4 8 4 8 4-2 4-8 4-8-4-8-4z"/>
                        <circle cx="9" cy="9" r="2.5"/>
                    </svg>
                </button>
            </div>
            <p class="template-description">${escapeHtml(tpl.Description || '无描述')}</p>
            <button class="btn btn-primary" onclick="openGenerateModal('${escapeHtml(tpl.Name)}')">
                <span>使用此模板</span>
                <svg width="16" height="16" viewBox="0 0 16 16" fill="none" stroke="currentColor">
                    <path d="M6 12l4-4-4-4"/>
                </svg>
            </button>
        `;
        container.appendChild(card);
    });
}

// Filter templates by search
function filterTemplates() {
    const searchInput = document.getElementById('search-input');
    const query = searchInput.value.toLowerCase().trim();
    
    if (!query) {
        renderTemplates(allTemplates);
        return;
    }
    
    const filtered = allTemplates.filter(tpl => {
        const name = (tpl.Name || '').toLowerCase();
        const desc = (tpl.Description || '').toLowerCase();
        return name.includes(query) || desc.includes(query);
    });
    
    renderTemplates(filtered);
}

// Open preview modal
async function openPreviewModal(templateName) {
    const modal = document.getElementById('preview-modal');
    document.getElementById('preview-title').textContent = `模板详情 - ${escapeHtml(templateName)}`;
    const content = document.getElementById('preview-content');
    content.innerHTML = '<div class="loading">加载中...</div>';
    
    try {
        const res = await fetch(`/api/templates/${encodeURIComponent(templateName)}`);
        if (!res.ok) throw new Error('Failed to load template details');
        
        const manifest = await res.json();
        const template = allTemplates.find(t => t.Name === templateName);
        
        let html = `
            <div class="preview-info">
                <div class="preview-item">
                    <strong>模板名称：</strong>
                    <span>${escapeHtml(templateName)}</span>
                </div>
                <div class="preview-item">
                    <strong>描述：</strong>
                    <span>${escapeHtml(template?.Description || manifest.description || '无描述')}</span>
                </div>
        `;
        
        if (manifest.fields && manifest.fields.length > 0) {
            html += `
                <div class="preview-item">
                    <strong>配置字段：</strong>
                    <div class="preview-fields">
            `;
            manifest.fields.forEach(field => {
                html += `
                    <div class="preview-field">
                        <div class="preview-field-name">
                            ${escapeHtml(field.prompt || field.name)}
                            ${field.required ? '<span class="required-badge">必填</span>' : ''}
                        </div>
                        ${field.description ? `<div class="preview-field-desc">${escapeHtml(field.description)}</div>` : ''}
                        ${field.default ? `<div class="preview-field-default">默认值: <code>${escapeHtml(field.default)}</code></div>` : ''}
                    </div>
                `;
            });
            html += `
                    </div>
                </div>
            `;
        } else {
            html += `
                <div class="preview-item">
                    <strong>配置字段：</strong>
                    <span style="color: var(--text-secondary);">此模板无需配置参数</span>
                </div>
            `;
        }
        
        html += '</div>';
        content.innerHTML = html;
        
        // 设置使用按钮
        document.getElementById('preview-use-btn').onclick = () => {
            closePreviewModal();
            openGenerateModal(templateName);
        };
        
        modal.classList.add('active');
    } catch (error) {
        content.innerHTML = `<div class="alert alert-error">加载失败: ${escapeHtml(error.message)}</div>`;
    }
}

// Close preview modal
function closePreviewModal() {
    document.getElementById('preview-modal').classList.remove('active');
}

// Escape HTML
function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}

// Toast Notification
function showToast(message, type = 'success') {
    const toast = document.createElement('div');
    toast.className = `toast toast-${type}`;
    const icon = type === 'success' ? '✅' : '❌';
    toast.innerHTML = `
        <span>${icon}</span>
        <span>${escapeHtml(message)}</span>
    `;
    document.body.appendChild(toast);
    
    setTimeout(() => {
        toast.style.animation = 'toastSlideIn 0.3s ease-out reverse';
        setTimeout(() => toast.remove(), 300);
    }, 3000);
}

// Initialize
loadTemplates();
