<template>
  <el-container>
    <el-main>
      <div v-if="articles && articles.length">
        <el-card v-for="article in articles" :key="article.ID" class="article-card">
          <h2>{{ article.Title }}</h2>
          <p>{{ article.Preview }}</p>
          <el-button text @click="viewDetail(article.ID)">阅读更多</el-button>
        </el-card>

        <!-- 分页条 -->
        <div class="pagination-bar">
          <el-pagination
            v-model:current-page="currentPage"
            :page-size="pageSize"
            :total="totalArticles"
            layout="prev, pager, next, jumper, total"
            background
            @current-change="handlePageChange"
          />
        </div>
      </div>
      <div v-else-if="loading" class="no-data">
        <el-skeleton :rows="5" animated />
      </div>
      <div v-else class="no-data">暂无文章，请先登录或联系管理员添加内容</div>
    </el-main>
  </el-container>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { useRouter } from 'vue-router';
import { ElMessage } from 'element-plus';
import axios from '../axios';
import { useAuthStore } from '../store/auth';
import type { Article } from "../types/Article";

const articles = ref<Article[]>([]);
const router = useRouter();
const authStore = useAuthStore();

// 分页状态
const currentPage = ref(1);
const pageSize = 20;
const totalArticles = ref(0);
const loading = ref(false);

// 响应类型
interface ArticlesResponse {
  source: string;
  articles: Article[];
  total: number;
  page: number;
  pageSize: number;
  totalPages: number;
}

const fetchArticles = async (page = 1) => {
  loading.value = true;
  try {
    const response = await axios.get<ArticlesResponse>('/articles', {
      params: { page, pageSize }
    });
    articles.value = response.data.articles ?? [];
    totalArticles.value = response.data.total ?? 0;
    currentPage.value = response.data.page ?? page;
  } catch (error) {
    console.error('Failed to load articles:', error);
    ElMessage.error('文章加载失败，请检查网络或重新登录');
  } finally {
    loading.value = false;
  }
};

const handlePageChange = (page: number) => {
  // 滚动到顶部提升体验
  window.scrollTo({ top: 0, behavior: 'smooth' });
  fetchArticles(page);
};

const viewDetail = (id: string) => {
  if (!authStore.isAuthenticated) {
    ElMessage.error('请先登录后再查看');
    return;
  }
  router.push({ name: 'NewsDetail', params: { id } });
};

onMounted(() => fetchArticles(1));
</script>

<style scoped>
.article-card {
  margin: 20px 0;
}

.pagination-bar {
  display: flex;
  justify-content: center;
  margin: 24px 0 16px;
}

.no-data {
  text-align: center;
  font-size: 1.2em;
  color: #999;
  margin-top: 40px;
}
</style>
