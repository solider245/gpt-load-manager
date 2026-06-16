<template>
  <div>
    <el-button text @click="$router.push('/servers')">
      <el-icon><ArrowLeft /></el-icon> 返回列表
    </el-button>

    <div v-if="server">
      <el-card shadow="never" style="margin-top:1rem">
        <template #header>
          <div style="display:flex;justify-content:space-between;align-items:center">
            <span>
              <el-tag :type="statusTag(server.status)" size="small" effect="dark" style="margin-right:8px">
                {{ server.status }}
              </el-tag>
              {{ server.name }}
            </span>
            <div>
              <el-button size="small" @click="$router.push(`/servers/${server.id}/edit`)">编辑</el-button>
              <el-button size="small" type="primary" @click="checkHealth" :loading="checking">
                <el-icon style="margin-right:4px"><Refresh /></el-icon> 检查健康状态
              </el-button>
            </div>
          </div>
        </template>

        <el-descriptions :column="2" border>
          <el-descriptions-item label="地址" :span="2">{{ server.host }}:{{ server.gpt_port }}</el-descriptions-item>
          <el-descriptions-item label="运行模式">{{ server.gpt_mode }}</el-descriptions-item>
          <el-descriptions-item label="版本">{{ server.version || '-' }}</el-descriptions-item>
          <el-descriptions-item label="SSH 端口">{{ server.ssh_port }}</el-descriptions-item>
          <el-descriptions-item label="SSH 认证方式">{{ server.auth_type }}</el-descriptions-item>
          <el-descriptions-item label="最后检查时间">{{ server.last_health_at ? new Date(server.last_health_at).toLocaleString() : '-' }}</el-descriptions-item>
          <el-descriptions-item label="创建时间">{{ new Date(server.created_at).toLocaleString() }}</el-descriptions-item>
        </el-descriptions>
      </el-card>

      <!-- Health check result -->
      <el-card v-if="healthResult" shadow="never" style="margin-top:1rem">
        <template #header>健康检查结果</template>
        <div v-if="healthResult.ssh" style="margin-bottom:1rem">
          <h4>SSH 连接</h4>
          <el-tag :type="healthResult.ssh.online ? 'success' : 'danger'" size="small">
            {{ healthResult.ssh.online ? '连通' : '失败' }}
          </el-tag>
          <pre v-if="healthResult.ssh.info" style="margin-top:8px;background:#f5f7fa;padding:8px;border-radius:4px">{{ JSON.stringify(healthResult.ssh.info, null, 2) }}</pre>
          <el-alert v-if="healthResult.ssh.error" :title="healthResult.ssh.error" type="error" show-icon style="margin-top:8px" />
        </div>
        <div>
          <h4>HTTP 健康检查</h4>
          <el-tag :type="healthResult.http?.online ? 'success' : 'danger'" size="small">
            {{ healthResult.http?.online ? `在线 (${healthResult.http.response_ms}ms)` : '离线' }}
          </el-tag>
          <pre style="margin-top:8px;background:#f5f7fa;padding:8px;border-radius:4px">{{ JSON.stringify(healthResult.http, null, 2) }}</pre>
        </div>
      </el-card>

      <!-- Deploy logs placeholder -->
      <el-card shadow="never" style="margin-top:1rem">
        <template #header>部署记录</template>
        <el-empty description="暂无部署记录" />
      </el-card>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import { ArrowLeft, Refresh } from '@element-plus/icons-vue'
import api, { handleError } from '@/api'

const route = useRoute()
const server = ref(null)
const healthResult = ref(null)
const checking = ref(false)

async function fetchServer() {
  try {
    const { data } = await api.getServer(route.params.id)
    server.value = data.data
  } catch (err) {
    ElMessage.error(handleError(err))
  }
}

async function checkHealth() {
  checking.value = true
  try {
    const { data } = await api.checkServerHealth(route.params.id)
    healthResult.value = data.data
  } catch (err) {
    ElMessage.error(handleError(err))
  } finally {
    checking.value = false
    await fetchServer()
  }
}

function statusTag(s) {
  return { online: 'success', offline: 'danger', unknown: 'info' }[s] || 'info'
}

onMounted(fetchServer)
</script>
