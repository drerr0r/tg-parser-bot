<template>
  <div class="rules">
    <el-row justify="space-between" align="middle" style="margin-bottom: 20px;">
      <el-col :span="12">
        <h2>Правила парсинга</h2>
      </el-col>
      <el-col :span="12" style="text-align: right;">
        <el-button type="primary" @click="showAddRule = true" icon="Plus">
          Добавить правило
        </el-button>
      </el-col>
    </el-row>

    <el-table :data="rules" v-loading="loading" empty-text="Нет правил">
      <el-table-column prop="id" label="ID" width="60" />
      <el-table-column prop="name" label="Название" />
      <el-table-column prop="source_channel" label="Канал" />
      <el-table-column prop="keywords" label="Ключевые слова">
        <template #default="scope">
          <span v-if="Array.isArray(scope.row.keywords)">{{ scope.row.keywords.join(', ') }}</span>
          <span v-else>{{ scope.row.keywords }}</span>
        </template>
      </el-table-column>
      <el-table-column prop="is_active" label="Статус">
        <template #default="scope">
          <el-tag v-if="scope.row.is_active" type="success">Активно</el-tag>
          <el-tag v-else type="danger">Неактивно</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="Действия" width="120">
        <template #default="scope">
          <el-button size="small" @click="editRule(scope.row)" icon="Edit" />
          <el-button size="small" type="danger" @click="deleteRuleHandler(scope.row.id)" icon="Delete" />
        </template>
      </el-table-column>
    </el-table>

    <!-- Полная форма -->
    <el-dialog v-model="showAddRule" :title="editingRule ? 'Редактировать правило' : 'Добавить правило'" width="600px">
      <el-form :model="ruleForm" label-width="160px">
        <el-form-item label="Название правила" required>
          <el-input v-model="ruleForm.name" placeholder="Мое правило" />
        </el-form-item>
        <el-form-item label="Канал источник" required>
          <el-input v-model="ruleForm.source_channel" placeholder="t.me/NewsWorldTrading" />
        </el-form-item>
        
        <el-form-item label="Ключевые слова">
          <el-input 
            v-model="ruleForm.keywords" 
            placeholder="новости,финансы,трейдинг,рынок"
            type="textarea"
            :rows="2"
          />
          <div class="form-help">Укажите через запятую</div>
        </el-form-item>

        <el-form-item label="Исключить слова">
          <el-input 
            v-model="ruleForm.exclude_words" 
            placeholder="реклама,спам,куплю,продам"
            type="textarea"
            :rows="2"
          />
          <div class="form-help">Укажите через запятую</div>
        </el-form-item>

        <el-form-item label="Типы медиа">
          <el-select v-model="ruleForm.media_types" multiple placeholder="Выберите типы">
            <el-option label="Текст" value="text" />
            <el-option label="Фото" value="photo" />
            <el-option label="Видео" value="video" />
            <el-option label="Документ" value="document" />
          </el-select>
        </el-form-item>

        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="Мин. длина текста">
              <el-input-number v-model="ruleForm.min_text_length" :min="0" :max="1000" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="Макс. длина текста">
              <el-input-number v-model="ruleForm.max_text_length" :min="1" :max="5000" />
            </el-form-item>
          </el-col>
        </el-row>

        <el-form-item label="Замены текста">
          <el-input 
            v-model="ruleForm.text_replacements" 
            placeholder="гарант:возможность, купить:рассмотреть"
            type="textarea"
            :rows="2"
          />
          <div class="form-help">Формат: слово:замена, слово:замена</div>
        </el-form-item>

        <el-form-item label="Префикс">
          <el-input v-model="ruleForm.add_prefix" placeholder="📈 " />
        </el-form-item>

        <el-form-item label="Суффикс">
          <el-input v-model="ruleForm.add_suffix" placeholder=" #финансы" />
        </el-form-item>

        <el-form-item label="Платформы публикации">
          <el-checkbox-group v-model="ruleForm.target_platforms">
            <el-checkbox label="telegram">Telegram</el-checkbox>
            <el-checkbox label="vk">VK</el-checkbox>
          </el-checkbox-group>
        </el-form-item>

        <el-form-item label="Активно">
          <el-switch v-model="ruleForm.is_active" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showAddRule = false">Отмена</el-button>
        <el-button type="primary" @click="saveRule">Сохранить</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script>
import { mapState, mapActions } from 'vuex'

export default {
  name: 'Rules',
  data() {
    return {
      showAddRule: false,
      editingRule: null,
      ruleForm: {
        name: '',
        source_channel: '',
        keywords: '',
        exclude_words: '',
        media_types: ['text', 'photo'],
        min_text_length: 10,
        max_text_length: 1000,
        text_replacements: '',
        add_prefix: '',
        add_suffix: '',
        target_platforms: ['telegram', 'vk'],
        is_active: true
      }
    }
  },
  computed: {
    ...mapState(['rules', 'loading'])
  },
  mounted() {
    this.fetchRules()
  },
  methods: {
    ...mapActions(['fetchRules', 'createRule', 'updateRule', 'deleteRule']),
    
    editRule(rule) {
      this.editingRule = rule
      
      // Преобразуем данные для формы
      this.ruleForm = { 
        name: rule.name || '',
        source_channel: rule.source_channel || '',
        keywords: Array.isArray(rule.keywords) ? rule.keywords.join(', ') : rule.keywords || '',
        exclude_words: Array.isArray(rule.exclude_words) ? rule.exclude_words.join(', ') : rule.exclude_words || '',
        media_types: Array.isArray(rule.media_types) ? rule.media_types : ['text', 'photo'],
        min_text_length: rule.min_text_length || 10,
        max_text_length: rule.max_text_length || 1000,
        text_replacements: this.formatTextReplacements(rule.text_replacements),
        add_prefix: rule.add_prefix || '',
        add_suffix: rule.add_suffix || '',
        target_platforms: Array.isArray(rule.target_platforms) ? rule.target_platforms : ['telegram', 'vk'],
        is_active: rule.is_active !== false
      }
      this.showAddRule = true
    },

    // Преобразуем объект замен в строку для формы
    formatTextReplacements(replacements) {
      if (!replacements || typeof replacements !== 'object') return ''
      return Object.entries(replacements)
        .map(([key, value]) => `${key}:${value}`)
        .join(', ')
    },

    // Преобразуем строку замен в объект для API
    parseTextReplacements(text) {
      if (!text.trim()) return {}
      
      const replacements = {}
      text.split(',').forEach(pair => {
        const [key, value] = pair.split(':').map(s => s.trim())
        if (key && value) {
          replacements[key] = value
        }
      })
      return replacements
    },

    async deleteRuleHandler(id) {
      try {
        await this.$confirm('Удалить правило?', 'Подтверждение', {
          type: 'warning'
        })
        await this.deleteRule(id)
        this.$message.success('Правило удалено')
      } catch (error) {
        if (error !== 'cancel') {
          this.$message.error('Ошибка при удалении правила')
        }
      }
    },

    async saveRule() {
      try {
        // Валидация
        if (!this.ruleForm.name.trim()) {
          this.$message.error('Введите название правила')
          return
        }
        if (!this.ruleForm.source_channel.trim()) {
          this.$message.error('Введите канал источник')
          return
        }

        // Подготавливаем данные для API
        const ruleData = {
          name: this.ruleForm.name.trim(),
          source_channel: this.ruleForm.source_channel.trim(),
          keywords: this.ruleForm.keywords ? 
            this.ruleForm.keywords.split(',').map(k => k.trim()).filter(k => k) : [],
          exclude_words: this.ruleForm.exclude_words ? 
            this.ruleForm.exclude_words.split(',').map(k => k.trim()).filter(k => k) : [],
          media_types: this.ruleForm.media_types,
          min_text_length: this.ruleForm.min_text_length,
          max_text_length: this.ruleForm.max_text_length,
          text_replacements: this.parseTextReplacements(this.ruleForm.text_replacements),
          add_prefix: this.ruleForm.add_prefix,
          add_suffix: this.ruleForm.add_suffix,
          target_platforms: this.ruleForm.target_platforms,
          is_active: this.ruleForm.is_active
        }

        console.log('Sending complete rule data:', ruleData)

        if (this.editingRule) {
          await this.updateRule({
            id: this.editingRule.id,
            ruleData: ruleData
          })
          this.$message.success('Правило обновлено')
        } else {
          await this.createRule(ruleData)
          this.$message.success('Правило создано')
        }
        this.showAddRule = false
        this.resetForm()
      } catch (error) {
        console.error('Save error details:', error)
        this.$message.error('Ошибка: ' + (error.response?.data?.error || error.message))
      }
    },

    resetForm() {
      this.editingRule = null
      this.ruleForm = {
        name: '',
        source_channel: '',
        keywords: '',
        exclude_words: '',
        media_types: ['text', 'photo'],
        min_text_length: 10,
        max_text_length: 1000,
        text_replacements: '',
        add_prefix: '',
        add_suffix: '',
        target_platforms: ['telegram', 'vk'],
        is_active: true
      }
    }
  }
}
</script>

<style scoped>
.rules {
  padding: 20px;
}

.form-help {
  font-size: 12px;
  color: #909399;
  margin-top: 4px;
}
</style>