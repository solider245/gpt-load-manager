<template>
  <div>
    <el-button text @click="$router.push('/servers')">
      <el-icon><ArrowLeft /></el-icon> 返回列表
    </el-button>
    <h2 style="margin-top:1rem">{{ isEdit ? '编辑服务器' : '添加服务器' }}</h2>

    <el-card shadow="never" style="max-width:600px;margin-top:1rem">
      <el-form :model="form" label-width="110px" @submit.prevent="save">
        <el-form-item label="名称" required>
          <el-input v-model="form.name" placeholder="例如: 新加坡-Master" />
        </el-form-item>
        <el-form-item label="IP 地址 / 域名" required>
          <el-input v-model="form.host" placeholder="IP 或域名" />
        </el-form-item>
        <el-form-item label="SSH 端口">
          <el-input-number v-model="form.ssh_port" :min="1" :max="65535" />
        </el-form-item>
        <el-form-item label="认证方式">
          <el-radio-group v-model="form.auth_type">
            <el-radio value="password">密码</el-radio>
            <el-radio value="key">私钥</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="凭据" required>
          <el-input
            v-if="form.auth_type === 'password'"
            v-model="form.auth_credential"
            type="password"
            placeholder="SSH 密码"
            show-password
          />
          <el-input
            v-else
            v-model="form.auth_credential"
            type="textarea"
            :rows="4"
            placeholder="SSH 私钥内容（支持 PEM 格式）"
          />
        </el-form-item>
        <el-form-item label="GPT 端口">
          <el-input-number v-model="form.gpt_port" :min="1" :max="65535" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" native-type="submit" :loading="saving">
            {{ isEdit ? '保存' : '添加' }}
          </el-button>
          <el-button @click="$router.push('/servers')">取消</el-button>
        </el-form-item>
      </el-form>
    </el-card>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { ArrowLeft } from '@element-plus/icons-vue'
import api, { handleError } from '@/api'

const route = useRoute()
const router = useRouter()
const isEdit = computed(() => !!route.params.id)
const saving = ref(false)

const form = ref({
  name: '',
  host: '',
  ssh_port: 22,
  auth_type: 'password',
  auth_credential: '',
  gpt_port: 3001,
})

onMounted(async () => {
  if (isEdit.value) {
    try {
      const { data } = await api.getServer(route.params.id)
      const s = data.data
      form.value = {
        name: s.name,
        host: s.host,
        ssh_port: s.ssh_port,
        auth_type: s.auth_type,
        auth_credential: '', // don't populate credential for security
        gpt_port: s.gpt_port,
      }
    } catch (err) {
      ElMessage.error(handleError(err))
    }
  }
})

async function save() {
  if (!form.value.name || !form.value.host) {
    ElMessage.warning('名称和地址不能为空')
    return
  }
  if (!isEdit.value && !form.value.auth_credential) {
    ElMessage.warning('请输入 SSH 凭据')
    return
  }

  saving.value = true
  try {
    if (isEdit.value) {
      const payload = { ...form.value }
      if (!payload.auth_credential) delete payload.auth_credential
      await api.updateServer(route.params.id, payload)
      ElMessage.success('已更新')
    } else {
      await api.createServer(form.value)
      ElMessage.success('已添加')
    }
    router.push('/servers')
  } catch (err) {
    ElMessage.error(handleError(err))
  } finally {
    saving.value = false
  }
}
</script>
