<template>
  <div id="app">
    <el-container>
      <el-header>
        <el-menu mode="horizontal" router :default-active="$route.path">
          <el-menu-item index="/">Главная</el-menu-item>
          <el-menu-item index="/rules">Правила</el-menu-item>
          <el-menu-item index="/posts">Посты</el-menu-item>
          <el-menu-item index="/stats">Статистика</el-menu-item>
        </el-menu>
      </el-header>
      <el-main>
        <router-view />
      </el-main>
    </el-container>
  </div>
</template>

<script>
export default {
  name: 'App'
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
</style>
