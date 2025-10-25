<template>
  <div class="posts">
    <h2>Обработанные посты</h2>

    <el-table :data="posts" v-loading="loading" empty-text="Нет постов">
      <el-table-column prop="id" label="ID" width="60" />
      <el-table-column prop="source_channel" label="Источник" />
      <el-table-column prop="content" label="Контент" :show-overflow-tooltip="true" />
      <el-table-column prop="published_telegram" label="Telegram">
        <template #default="scope">
          <el-tag v-if="scope.row.published_telegram" type="success">Да</el-tag>
          <el-tag v-else type="info">Нет</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="published_vk" label="VK">
        <template #default="scope">
          <el-tag v-if="scope.row.published_vk" type="success">Да</el-tag>
          <el-tag v-else type="info">Нет</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="parsed_at" label="Дата" width="180">
        <template #default="scope">
          {{ formatDate(scope.row.parsed_at) }}
        </template>
      </el-table-column>
    </el-table>
  </div>
</template>

<script>
import { mapState, mapActions } from 'vuex'

export default {
  name: 'Posts',
  computed: {
    ...mapState(['posts', 'loading'])
  },
  mounted() {
    this.fetchPosts()
  },
  methods: {
    ...mapActions(['fetchPosts']),
    formatDate(dateString) {
      if (!dateString) return 'Нет данных'
      
      try {
        const date = new Date(dateString)
        return date.toLocaleString('ru-RU', {
          year: 'numeric',
          month: '2-digit',
          day: '2-digit',
          hour: '2-digit',
          minute: '2-digit',
          second: '2-digit'
        })
      } catch (error) {
        console.error('Error formatting date:', error)
        return 'Неверная дата'
      }
    }
  }
}
</script>

<style scoped>
.posts {
  padding: 20px;
}
</style>