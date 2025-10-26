<template>
  <div id="app">
    <el-container v-if="isAuthenticated">
      <el-header>
        <el-menu mode="horizontal" router :default-active="$route.path">
          <el-menu-item index="/">Главная</el-menu-item>
          <el-menu-item index="/rules">Правила</el-menu-item>
          <el-menu-item index="/posts">Посты</el-menu-item>
          <el-menu-item index="/stats">Статистика</el-menu-item>
          <el-menu-item index="/logs">Логи</el-menu-item>
          <el-sub-menu index="user" class="user-menu">
            <template #title>
              <el-avatar :size="30" :src="userAvatar" style="margin-right: 8px;" />
              {{ user?.username }}
            </template>
            <el-menu-item @click="handleLogout">
              <el-icon><SwitchButton /></el-icon>
              Выйти
            </el-menu-item>
          </el-sub-menu>
        </el-menu>
      </el-header>
      <el-main>
        <router-view />
      </el-main>
    </el-container>
    
    <router-view v-else />
  </div>
</template>

<script>
import { mapState, mapActions } from 'vuex'
import { SwitchButton } from '@element-plus/icons-vue'

export default {
  name: 'App',
  components: {
    SwitchButton
  },
  computed: {
    ...mapState(['isAuthenticated', 'user']),
    userAvatar() {
      return ''
    }
  },
  methods: {
    ...mapActions(['checkAuth', 'logout']),
    
    async handleLogout() {
      try {
        await this.$confirm('Вы уверены, что хотите выйти?', 'Подтверждение выхода', {
          type: 'warning'
        })
        this.logout()
        this.$message.success('Вы успешно вышли из системы')
      } catch (error) {
        if (error !== 'cancel') {
          console.error('Logout error:', error)
        }
      }
    }
  }
}
</script>

<style>
#app {
  font-family: Avenir, Helvetica, Arial, sans-serif;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
  color: #2c3e50;
}

.el-header {
  padding: 0;
  border-bottom: 1px solid #dcdfe6;
}

.user-menu {
  float: right;
  margin-right: 20px;
}
</style>