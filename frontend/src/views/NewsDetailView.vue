<template>
  <el-container>
    <el-main>
      <el-card v-if="article" class="article-detail">
        <h1>{{ article.Title }}</h1>
        <p class="article-content">{{ article.Content }}</p>

        <div class="like-section">
          <el-button type="primary" @click="likeArticle">点赞</el-button>
          <p>点赞数: {{ likes }}</p>
        </div>

        <!-- ── AI 工具栏 ── -->
        <el-divider content-position="left">
          <span class="ai-label">🤖 AI 助手</span>
        </el-divider>

        <div class="ai-toolbar">
          <el-button
            type="success"
            :loading="loading && activeAction === 'summarize'"
            :disabled="loading"
            @click="runAI('summarize')"
          >
            📝 生成摘要
          </el-button>
          <el-button
            type="warning"
            :loading="loading && activeAction === 'translate'"
            :disabled="loading"
            @click="runAI('translate')"
          >
            🌐 翻译文章
          </el-button>
          <el-button
            v-if="summarizeResult || translateResult"
            text
            type="info"
            @click="clearResults"
          >
            清除结果
          </el-button>
        </div>

        <!-- AI 结果展示区 -->
        <div v-if="summarizeResult || translateResult || loading" class="ai-result-area">
          <el-tabs v-model="activeTab">
            <el-tab-pane
              v-if="summarizeResult || (loading && activeAction === 'summarize')"
              label="📝 摘要"
              name="summarize"
            >
              <div class="ai-result-text">
                <span v-if="loading && activeAction === 'summarize' && !summarizeResult">
                  <el-skeleton :rows="3" animated />
                </span>
                <span v-else>{{ summarizeResult }}</span>
                <span v-if="loading && activeAction === 'summarize'" class="cursor-blink">▌</span>
              </div>
            </el-tab-pane>

            <el-tab-pane
              v-if="translateResult || (loading && activeAction === 'translate')"
              label="🌐 翻译"
              name="translate"
            >
              <div class="ai-result-text">
                <span v-if="loading && activeAction === 'translate' && !translateResult">
                  <el-skeleton :rows="5" animated />
                </span>
                <span v-else>{{ translateResult }}</span>
                <span v-if="loading && activeAction === 'translate'" class="cursor-blink">▌</span>
              </div>
            </el-tab-pane>
          </el-tabs>
        </div>
      </el-card>

      <div v-else class="no-data">您必须登录/注册才可以阅读文章</div>
    </el-main>
  </el-container>
</template>

<script setup lang="ts">
import { ref, onMounted } from "vue";
import { useRoute } from "vue-router";
import { ElMessage } from "element-plus";
import axios from "../axios";
import type { Article, Like } from "../types/Article";

const article = ref<Article | null>(null);
const route = useRoute();
const likes = ref<number>(0);

// AI 相关状态
const loading = ref(false);
const activeAction = ref<"summarize" | "translate">("summarize");
const activeTab = ref<"summarize" | "translate">("summarize");
const summarizeResult = ref("");
const translateResult = ref("");
let currentEventSource: EventSource | null = null;

const { id } = route.params;

const fetchArticle = async () => {
  try {
    const response = await axios.get<Article>(`/articles/${id}`);
    article.value = response.data;
  } catch (error) {
    console.error("Failed to load article:", error);
  }
};

const likeArticle = async () => {
  try {
    const res = await axios.post<Like>(`articles/${id}/like`);
    likes.value = res.data.likes;
    await fetchLike();
  } catch (error) {
    console.log("Error Liking article:", error);
  }
};

const fetchLike = async () => {
  try {
    const res = await axios.get<Like>(`articles/${id}/like`);
    likes.value = res.data.likes;
  } catch (error) {
    console.log("Error fetching likes:", error);
  }
};

// ── AI 流式请求 ──
const runAI = (action: "summarize" | "translate") => {
  // 如果当前有请求，先关闭
  if (currentEventSource) {
    currentEventSource.close();
    currentEventSource = null;
  }

  // 清空对应结果
  if (action === "summarize") {
    summarizeResult.value = "";
  } else {
    translateResult.value = "";
  }

  activeAction.value = action;
  activeTab.value = action;
  loading.value = true;

  // 从 localStorage 取 token，手动拼入 URL（SSE 不支持自定义请求头）
  const token = localStorage.getItem("token") ?? "";
  const url = `http://localhost:3000/api/articles/${id}/ai?action=${action}&token=${encodeURIComponent(token)}`;

  const es = new EventSource(url);
  currentEventSource = es;

  es.onmessage = (event) => {
    if (event.data === "[DONE]") {
      loading.value = false;
      es.close();
      currentEventSource = null;
      return;
    }
    if (action === "summarize") {
      summarizeResult.value += event.data;
    } else {
      translateResult.value += event.data;
    }
  };

  es.onerror = (err) => {
    console.error("SSE error:", err);
    loading.value = false;
    es.close();
    currentEventSource = null;
    ElMessage.error("AI 请求失败，请稍后重试");
  };
};

const clearResults = () => {
  summarizeResult.value = "";
  translateResult.value = "";
};

onMounted(fetchArticle);
onMounted(fetchLike);
</script>

<style scoped>
.article-detail {
  margin: 20px 0;
}

.article-content {
  white-space: pre-wrap;
  line-height: 1.8;
  color: #333;
}

.like-section {
  display: flex;
  align-items: center;
  gap: 12px;
  margin: 16px 0;
}

.like-section p {
  margin: 0;
  color: #666;
}

.ai-label {
  font-size: 15px;
  font-weight: 600;
  color: #409eff;
}

.ai-toolbar {
  display: flex;
  gap: 12px;
  flex-wrap: wrap;
  margin-bottom: 16px;
}

.ai-result-area {
  background: #f8f9fa;
  border-radius: 8px;
  padding: 16px;
  margin-top: 8px;
  border: 1px solid #e4e7ed;
}

.ai-result-text {
  white-space: pre-wrap;
  line-height: 1.9;
  color: #2c3e50;
  font-size: 14px;
  min-height: 60px;
}

.cursor-blink {
  display: inline-block;
  animation: blink 0.8s step-end infinite;
  color: #409eff;
  font-weight: bold;
}

@keyframes blink {
  0%, 100% { opacity: 1; }
  50%       { opacity: 0; }
}

.no-data {
  text-align: center;
  font-size: 1.2em;
  color: #999;
}
</style>
