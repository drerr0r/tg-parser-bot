<template>
  <div class="logs">
    <h2>Логи системы</h2>

    <el-card>
      <template #header>
        <div class="logs-header">
          <span>Журнал событий</span>
          <div>
            <el-button @click="clearFilters" :disabled="loading">Сбросить фильтры</el-button>
            <el-button type="primary" @click="refreshLogs" :loading="loading" icon="Refresh">
              Обновить
            </el-button>
          </div>
        </div>
      </template>

      <!-- Фильтры -->
      <div class="filters">
        <el-select v-model="filters.level" placeholder="Уровень лога" clearable @change="refreshLogs">
          <el-option label="INFO" value="info" />
          <el-option label="WARN" value="warn" />
          <el-option label="ERROR" value="error" />
          <el-option label="DEBUG" value="debug" />
        </el-select>

        <el-input
          v-model="filters.search"
          placeholder="Поиск по сообщению..."
          clearable
          @input="refreshLogs"
          style="width: 300px; margin-left: 10px;"
        >
          <template #prefix>
            <el-icon><Search /></el-icon>
          </template>
        </el-input>
      </div>

      <!-- Таблица логов -->
      <el-table 
        :data="logs" 
        v-loading="loading" 
        empty-text="Нет записей логов"
        style="margin-top: 20px;"
      >
        <el-table-column prop="timestamp" label="Время" width="180">
          <template #default="scope">
            {{ formatDateTime(scope.row.timestamp) }}
          </template>
        </el-table-column>
        
        <el-table-column prop="level" label="Уровень" width="100">
          <template #default="scope">
            <el-tag :type="getLevelType(scope.row.level)" size="small">
              {{ scope.row.level?.toUpperCase() }}
            </el-tag>
          </template>
        </el-table-column>
        
        <el-table-column prop="service" label="Сервис" width="150" />
        
        <el-table-column prop="message" label="Сообщение" min-width="300">
          <template #default="scope">
            <span :class="getMessageClass(scope.row.level)">{{ scope.row.message }}</span>
          </template>
        </el-table-column>
      </el-table>

      <!-- Пагинация -->
      <div class="pagination-container">
        <el-pagination
          v-model:current-page="pagination.page"
          v-model:page-size="pagination.limit"
          :page-sizes="[10, 20, 50, 100]"
          :total="pagination.total"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="refreshLogs"
          @current-change="refreshLogs"
        />
      </div>
    </el-card>
  </div>
</template>

<script>
import { Search } from '@element-plus/icons-vue'

export default {
  name: 'Logs',
  components: {
    Search
  },
  data() {
    return {
      logs: [],
      loading: false,
      filters: {
        level: '',
        search: ''
      },
      pagination: {
        page: 1,
        limit: 20,
        total: 0
      }
    }
  },
  mounted() {
    this.refreshLogs()
  },
  methods: {
async refreshLogs() {
  this.loading = true
  try {
    // Используем реальный API с параметрами
    await this.$store.dispatch('fetchLogs', {
      limit: this.pagination.limit,
      offset: (this.pagination.page - 1) * this.pagination.limit,
      level: this.filters.level,
      search: this.filters.search
    })
    
    // Получаем данные из store - ОБРАТИТЕ ВНИМАНИЕ НА ФОРМАТ!
    const logsData = this.$store.state.logs
    
    console.log('Данные логов из store:', logsData) // ДЛЯ ОТЛАДКИ
    
    // API возвращает { logs: [], total: number, limit: number, offset: number }
    if (logsData && logsData.logs && Array.isArray(logsData.logs)) {
      this.logs = logsData.logs
      this.pagination.total = logsData.total || logsData.logs.length
    } else if (Array.isArray(logsData)) {
      // Если по какой-то причине пришел просто массив (старый формат)
      this.logs = logsData
      this.pagination.total = logsData.length
    } else {
      console.error('Неверный формат данных логов:', logsData)
      this.logs = []
      this.pagination.total = 0
      this.$message.error('Ошибка формата данных логов')
    }
    
  } catch (error) {
    console.error('Ошибка загрузки логов:', error)
    // Если API не работает, используем заглушку
    this.logs = this.generateMockLogs()
    this.pagination.total = this.logs.length
    this.$message.warning('Используются тестовые данные логов')
  } finally {
    this.loading = false
  }
},
    clearFilters() {
      this.filters = {
        level: '',
        search: ''
      }
      this.pagination.page = 1
      this.refreshLogs()
    },

    formatDateTime(timestamp) {
      if (!timestamp) return 'Нет данных'
      try {
        const date = new Date(timestamp)
        return date.toLocaleString('ru-RU', {
          year: 'numeric',
          month: '2-digit',
          day: '2-digit',
          hour: '2-digit',
          minute: '2-digit',
          second: '2-digit'
        })
      } catch (error) {
        return 'Неверная дата'
      }
    },

    getLevelType(level) {
      const levelMap = {
        'error': 'danger',
        'warn': 'warning',
        'info': 'info',
        'debug': ''
      }
      return levelMap[level?.toLowerCase()] || 'info'
    },

    getMessageClass(level) {
      return {
        'error-message': level?.toLowerCase() === 'error',
        'warn-message': level?.toLowerCase() === 'warn'
      }
    },

    generateMockLogs() {
      const services = ['parser', 'publisher', 'api', 'database', 'auth']
      const levels = ['info', 'warn', 'error', 'debug']
      const messages = [
        'Запуск парсера Telegram',
        'Найдено новое сообщение в канале',
        'Ошибка подключения к базе данных',
        'Успешная публикация в Telegram',
        'Пользователь admin вошел в систему',
        'Создано новое правило парсинга',
        'Ошибка валидации правила',
        'Парсер завершил работу',
        'Получен запрос к API /api/rules',
        'Ошибка аутентификации: неверный токен'
      ]

      const logs = []
      const now = new Date()

      for (let i = 0; i < 50; i++) {
        const timestamp = new Date(now.getTime() - Math.random() * 24 * 60 * 60 * 1000)
        const level = levels[Math.floor(Math.random() * levels.length)]
        const service = services[Math.floor(Math.random() * services.length)]
        const message = messages[Math.floor(Math.random() * messages.length)]

        logs.push({
          id: i + 1,
          timestamp: timestamp.toISOString(),
          level: level,
          service: service,
          message: `${message} [ID: ${Math.floor(Math.random() * 1000)}]`
        })
      }

      // Сортируем по времени (новые сверху)
      return logs.sort((a, b) => new Date(b.timestamp) - new Date(a.timestamp))
    }
  }
}
</script>

<style scoped>
.logs {
  padding: 20px;
}

.logs-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.filters {
  margin-bottom: 20px;
}

.pagination-container {
  margin-top: 20px;
  display: flex;
  justify-content: center;
}

.error-message {
  color: #f56c6c;
  font-weight: 500;
}

.warn-message {
  color: #e6a23c;
  font-weight: 500;
}
</style>