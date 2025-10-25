<template>
  <div class="stats">
    <h2>Статистика</h2>

    <el-row :gutter="20" style="margin-bottom: 30px;">
      <el-col :span="6">
        <el-statistic title="Всего правил" :value="stats.rules_count || 0" />
      </el-col>
      <el-col :span="6">
        <el-statistic title="Всего постов" :value="stats.posts_count || 0" />
      </el-col>
      <el-col :span="6">
        <el-statistic title="Telegram постов" :value="stats.telegram_posts || 0" />
      </el-col>
      <el-col :span="6">
        <el-statistic title="VK постов" :value="stats.vk_posts || 0" />
      </el-col>
    </el-row>

    <el-card>
      <template #header>
        <span>Детальная статистика</span>
      </template>
      <el-descriptions :column="2" border v-if="stats">
        <el-descriptions-item label="Активных правил">
          {{ stats.active_rules || 0 }}
        </el-descriptions-item>
        <el-descriptions-item label="Неактивных правил">
          {{ stats.inactive_rules || 0 }}
        </el-descriptions-item>
        <el-descriptions-item label="Успешных публикаций">
          {{ stats.success_posts || 0 }}
        </el-descriptions-item>
        <el-descriptions-item label="Ошибок публикации">
          {{ stats.failed_posts || 0 }}
        </el-descriptions-item>
      </el-descriptions>
      <div v-else>
        <el-empty description="Нет данных статистики" />
      </div>
    </el-card>
  </div>
</template>

<script>
import { mapState, mapActions } from 'vuex'

export default {
  name: 'Stats',
  computed: {
    ...mapState(['stats', 'loading'])
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
.stats {
  padding: 20px;
}
</style>
