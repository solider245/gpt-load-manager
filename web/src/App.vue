<template>
  <el-container style="height: 100vh">
    <el-header>
      <div class="header-inner">
        <div class="header-left">
          <h1>GPT-Load Manager</h1>
          <el-menu mode="horizontal" router :ellipsis="false" :default-active="route.path">
            <el-menu-item index="/servers">
              <el-icon><Monitor /></el-icon>
              服务器
            </el-menu-item>
          </el-menu>
        </div>
        <div class="header-right">
          <span class="version-badge">v0.1.0</span>
        </div>
      </div>
    </el-header>
    <el-main>
      <router-view />
    </el-main>
  </el-container>
</template>

<script setup>
import { Monitor } from '@element-plus/icons-vue'
import { useRoute } from 'vue-router'
const route = useRoute()
</script>

<style>
body { margin: 0; font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; }
.header-inner {
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 100%;
}
.header-left {
  display: flex;
  align-items: center;
  gap: 2rem;
}
.header-left h1 {
  font-size: 1.2rem;
  white-space: nowrap;
  margin: 0;
}
.version-badge {
  font-size: 0.75rem;
  color: #999;
}
.el-header {
  box-shadow: 0 1px 4px rgba(0,0,0,.08);
  z-index: 10;
}
.el-main {
  background: #f5f7fa;
}
</style>
