<template>
  <div>
    <div class="page-header">
      <h2>服务器管理</h2>
      <el-button type="primary" @click="$router.push('/servers/add')">
        <el-icon><Plus /></el-icon> 添加服务器
      </el-button>
    </div>

    <el-card shadow="never">
      <el-table :data="servers" v-loading="loading" stripe style="width:100%">
        <el-table-column prop="name" label="名称" width="160" />
        <el-table-column label="地址" width="200">
          <template #default="{ row }">
            {{ row.host }}:{{ row.gpt_port }}
          </template>
        </el-table-column>
        <el-table-column prop="gpt_mode" label="模式" width="100" />
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="statusTag(row.status)" size="small" effect="dark">
              {{ row.status }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="version" label="版本" width="120" />
        <el-table-column label="最后健康检查" width="180">
          <template #default="{ row }">
            <span style="font-size:0.85rem;color:#999">
              {{ row.last_health_at ? new Date(row.last_health_at).toLocaleString() : '-' }}
            </span>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="260" fixed="right">
          <template #default="{ row }">
            <el-button size="small" @click="checkHealth(row)" :loading="checkingId === row.id">
              <el-icon style="margin-right:4px"><Refresh /></el-icon>检查
            </el-button>
            <el-button size="small" @click="$router.push(`/servers/${row.id}`)">详情</el-button>
            <el-popconfirm title="确定删除此服务器?" @confirm="removeServer(row.id)">
              <template #reference>
                <el-button size="small" type="danger">删除</el-button>
              </template>
            </el-popconfirm>
          </template>
        </el-table-column>
      </el-table>

      <el-empty v-if="!loading && servers.length === 0" description="还没有添加服务器" />
    </el-card>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { Plus, Refresh } from '@element-plus/icons-vue'
import api, { handleError } from '@/api'

const servers = ref([])
const loading = ref(false)
const checkingId = ref(null)

async function fetchServers() {
  loading.value = true
  try {
    const { data } = await api.listServers()
    servers.value = data.data || []
  } catch (err) {
    ElMessage.error(handleError(err))
  } finally {
    loading.value = false
  }
}

async function checkHealth(server) {
  checkingId.value = server.id
  try {
    const { data } = await api.checkServerHealth(server.id)
    const httpStatus = data.data?.http
    const sshStatus = data.data?.ssh

    if (httpStatus?.online) {
      ElMessage.success(`${server.name} 在线 (${httpStatus.response_ms}ms)`)
    } else {
      ElMessage.warning(`${server.name} 离线: ${httpStatus?.error || '无响应'}`)
    }
    if (sshStatus?.online) {
      ElMessage.info(`SSH 连通: ${sshStatus?.info?.hostname || ''}`)
    } else {
      ElMessage.warning(`SSH 连接失败: ${sshStatus?.error || ''}`)
    }
  } catch (err) {
    ElMessage.error(handleError(err))
  } finally {
    checkingId.value = null
    await fetchServers()
  }
}

async function removeServer(id) {
  try {
    await api.deleteServer(id)
    ElMessage.success('已删除')
    await fetchServers()
  } catch (err) {
    ElMessage.error(handleError(err))
  }
}

function statusTag(s) {
  return { online: 'success', offline: 'danger', unknown: 'info' }[s] || 'info'
}

onMounted(fetchServers)
</script>

<style scoped>
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1rem;
}
.page-header h2 { margin: 0; }
</style>
