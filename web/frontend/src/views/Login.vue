<template>
  <div class="login-container">
    <el-card class="login-card">
      <template #header>
        <div class="login-header">
          <h2>Вход в систему</h2>
          <p>TG Parser Bot - Панель управления</p>
        </div>
      </template>

      <el-form 
        :model="loginForm" 
        :rules="loginRules" 
        ref="loginFormRef"
        @submit.prevent="handleLogin"
      >
        <el-form-item prop="username">
          <el-input
            v-model="loginForm.username"
            placeholder="Имя пользователя"
            size="large"
            prefix-icon="User"
          />
        </el-form-item>

        <el-form-item prop="password">
          <el-input
            v-model="loginForm.password"
            type="password"
            placeholder="Пароль"
            size="large"
            prefix-icon="Lock"
            show-password
          />
        </el-form-item>

        <el-form-item>
          <el-button
            type="primary"
            size="large"
            :loading="loading"
            @click="handleLogin"
            style="width: 100%"
          >
            Войти
          </el-button>
        </el-form-item>
      </el-form>

      <div v-if="error" class="error-message">
        <el-alert
          :title="error"
          type="error"
          show-icon
          :closable="false"
        />
      </div>
    </el-card>
  </div>
</template>

<script>
import { mapActions, mapState } from 'vuex'

export default {
  name: 'Login',
  data() {
    return {
      loginForm: {
        username: '',
        password: ''
      },
      loginRules: {
        username: [
          { required: true, message: 'Введите имя пользователя', trigger: 'blur' }
        ],
        password: [
          { required: true, message: 'Введите пароль', trigger: 'blur' }
        ]
      },
      error: ''
    }
  },
  computed: {
    ...mapState(['loading'])
  },
  mounted() {
    // Если уже авторизован, перенаправляем на главную
    if (this.$store.state.isAuthenticated) {
      this.$router.push('/')
    }
  },
  methods: {
    ...mapActions(['login']),

    async handleLogin() {
      try {
        this.error = ''
        await this.$refs.loginFormRef.validate()
        
        await this.login(this.loginForm)
        this.$message.success('Успешный вход!')
        this.$router.push('/')
      } catch (error) {
        if (error.response?.data?.error) {
          this.error = error.response.data.error
        } else if (error.message) {
          this.error = error.message
        } else {
          this.error = 'Ошибка входа. Проверьте данные и попробуйте снова.'
        }
      }
    }
  }
}
</script>

<style scoped>
.login-container {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 100vh;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
}

.login-card {
  width: 400px;
  max-width: 90vw;
}

.login-header {
  text-align: center;
}

.login-header h2 {
  margin: 0 0 8px 0;
  color: #303133;
}

.login-header p {
  margin: 0;
  color: #909399;
  font-size: 14px;
}

.error-message {
  margin-top: 16px;
}
</style>