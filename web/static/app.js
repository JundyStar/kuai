let currentTemplate = null;
let modal = null;

function showTab(tab) {
    document.getElementById('panel-templates').style.display = tab === 'templates' ? 'block' : 'none';
    document.getElementById('panel-upload').style.display = tab === 'upload' ? 'block' : 'none';
    
    const links = document.querySelectorAll('.nav-link');
    links.forEach(link => link.classList.remove('active'));
    event.target.classList.add('active');
}

async function loadTemplates() {
    try {
        const res = await fetch('/api/templates');
        const templates = await res.json();
        const container = document.getElementById('templates-list');
        container.innerHTML = '';
        
        if (templates.length === 0) {
            container.innerHTML = '<p class="text-muted">暂无模板，请先上传模板</p>';
            return;
        }
        
        templates.forEach(tpl => {
            const col = document.createElement('div');
            col.className = 'col-md-6 col-lg-4 mb-4';
            col.innerHTML = `
                <div class="card template-card h-100">
                    <div class="card-body">
                        <h5 class="card-title">${tpl.Name}</h5>
                        <p class="card-text text-muted">${tpl.Description || '无描述'}</p>
                        <button class="btn btn-primary w-100" onclick="openGenerateModal('${tpl.Name}')">使用此模板</button>
                    </div>
                </div>
            `;
            container.appendChild(col);
        });
    } catch (error) {
        alert('加载模板失败: ' + error.message);
    }
}

async function openGenerateModal(templateName) {
    currentTemplate = templateName;
    document.getElementById('modal-title').textContent = '生成项目 - ' + templateName;
    
    try {
        const res = await fetch('/api/templates/' + templateName);
        const manifest = await res.json();
        
        const formFields = document.getElementById('form-fields');
        formFields.innerHTML = '';
        
        if (manifest.fields && manifest.fields.length > 0) {
            manifest.fields.forEach(field => {
                const div = document.createElement('div');
                div.className = 'mb-3';
                div.innerHTML = `
                    <label class="form-label">
                        ${field.prompt || field.name}
                        ${field.description ? ' <small class="text-muted">(' + field.description + ')</small>' : ''}
                        ${field.required ? '<span class="text-danger">*</span>' : ''}
                    </label>
                    <input 
                        type="text" 
                        name="${field.name}" 
                        class="form-control"
                        value="${field.default || ''}"
                        ${field.required ? 'required' : ''}
                        placeholder="${field.default ? '默认: ' + field.default : ''}"
                    >
                `;
                formFields.appendChild(div);
            });
        } else {
            formFields.innerHTML = '<p class="text-muted">此模板无需配置参数</p>';
        }
        
        modal = new bootstrap.Modal(document.getElementById('generate-modal'));
        modal.show();
    } catch (error) {
        alert('加载模板详情失败: ' + error.message);
    }
}

function closeModal() {
    if (modal) {
        modal.hide();
    }
    currentTemplate = null;
}

async function submitGenerate() {
    const form = document.getElementById('generate-form');
    const formData = new FormData(form);
    const values = {};
    formData.forEach((value, key) => {
        values[key] = value;
    });
    
    const messageDiv = document.getElementById('generate-message');
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
            window.location.href = result.downloadUrl;
            setTimeout(() => {
                closeModal();
                messageDiv.innerHTML = '';
            }, 2000);
        } else {
            messageDiv.innerHTML = '<div class="alert alert-danger">生成失败: ' + (result.message || '未知错误') + '</div>';
        }
    } catch (error) {
        messageDiv.innerHTML = '<div class="alert alert-danger">生成失败: ' + error.message + '</div>';
    }
}

document.getElementById('upload-form').addEventListener('submit', async (e) => {
    e.preventDefault();
    const formData = new FormData(e.target);
    const messageDiv = document.getElementById('upload-message');
    messageDiv.innerHTML = '<div class="alert alert-info">正在上传...</div>';
    
    try {
        const res = await fetch('/api/upload', {
            method: 'POST',
            body: formData
        });
        
        const result = await res.json();
        if (result.status === 'success') {
            messageDiv.innerHTML = '<div class="alert alert-success">✅ ' + result.message + '</div>';
            e.target.reset();
            loadTemplates();
        } else {
            messageDiv.innerHTML = '<div class="alert alert-danger">上传失败: ' + (result.message || '未知错误') + '</div>';
        }
    } catch (error) {
        messageDiv.innerHTML = '<div class="alert alert-danger">上传失败: ' + error.message + '</div>';
    }
});

loadTemplates();

