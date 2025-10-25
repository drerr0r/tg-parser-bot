<template>
  <div class="home">
    <h2>TG Parser Bot - Панель управления</h2>
    
    <el-row :gutter="20" style="margin-bottom: 20px;">
      <el-col :span="6">
        <el-statistic title="Правил" :value="stats.rules_count || 0" />
      </el-col>
      <el-col :span="6">
        <el-statistic title="Постов" :value="stats.posts_count || 0" />
      </el-col>
      <el-col :span="6">
        <el-statistic title="Telegram" :value="stats.telegram_posts || 0" />
      </el-col>
      <el-col :span="6">
        <el-statistic title="VK" :value="stats.vk_posts || 0" />
      </el-col>
    </el-row>

    <el-divider />

    <el-row :gutter="20">
      <el-col :span="12">
        <el-card>
          <template #header>
            <span>Быстрые действия</span>
          </template>
          <el-button type="primary" @click="$router.push('/rules')" icon="Plus">
            Добавить правило
          </el-button>
          <el-button @click="$router.push('/posts')" icon="List">
            Просмотр постов
          </el-button>
        </el-card>
      </el-col>
      <el-col :span="12">
        <el-card>
          <template #header>
            <span>Статус системы</span>
          </template>
          <el-tag type="success" style="margin-right: 10px;">Парсер активен</el-tag>
          <el-tag type="success" style="margin-right: 10px;">База данных</el-tag>
          <el-tag type="success" style="margin-right: 10px;">API</el-tag>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script>
import { mapState, mapActions } from 'vuex'

export default {
  name: 'Home',
  computed: {
    ...mapState(['stats'])
  },
  mounted() {
    this.fetchStats()
  },
  methods: {
    ...mapActions(['fetchStats'])
  }
}
</script>

<style scoped>
.home {
  padding: 20px;
}
</style>
