<style scoped>
.tts-container {
  width: 100%;
  height: 90vh;
  padding: 0;
  box-sizing: border-box;
  background: #f5f7fa;
}

.main-content {
  width: 100%;
  height: 100%;
  margin: 0 0;
}

.card {
  width: 100%;
  height: 100%;
  background: #fff;
  border-radius: 8px;
  padding: 20px;
  box-sizing: border-box;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.08);
  overflow: auto;
}

.card-title {
  font-size: 14px;
  font-weight: 600;
  color: #303133;
  margin-bottom: 16px;
  padding-bottom: 8px;
  border-bottom: 1px solid #ebeef5;
}

.form-row {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 12px;
}

.form-label {
  font-size: 13px;
  color: #606266;
  min-width: 80px;
}

.textarea-wrapper {
  margin-bottom: 16px;
}

.text-input {
  width: 100%;
  min-height: 150px;
  padding: 12px;
  border: 1px solid #dcdfe6;
  border-radius: 4px;
  font-size: 14px;
  resize: vertical;
  box-sizing: border-box;
  font-family: inherit;
}

.text-input:focus {
  outline: none;
  border-color: #409eff;
  box-shadow: 0 0 0 2px rgba(64, 158, 255, 0.1);
}

.api-key-input {
  width: 100%;
  padding: 8px 12px;
  border: 1px solid #dcdfe6;
  border-radius: 4px;
  font-size: 13px;
  box-sizing: border-box;
}

.api-key-input:focus {
  outline: none;
  border-color: #409eff;
}

.char-count {
  font-size: 12px;
  color: #909399;
  text-align: right;
  margin-top: 4px;
}

.submit-btn {
  width: 100%;
  padding: 10px 20px;
  background: #67c23a;
  border: none;
  border-radius: 4px;
  color: #fff;
  font-size: 14px;
  cursor: pointer;
  transition: background 0.2s;
}

.submit-btn:hover {
  background: #85ce61;
}

.submit-btn:disabled {
  background: #c0c4cc;
  cursor: not-allowed;
}

.result-card {
  margin-top: 16px;
  padding: 16px;
  border-radius: 6px;
  background: #f0f9eb;
  border: 1px solid #67c23a;
}

.result-card.error {
  background: #fef0f0;
  border-color: #f56c6c;
}

.result-title {
  font-size: 14px;
  font-weight: 600;
  margin-bottom: 8px;
}

.result-info {
  font-size: 13px;
  color: #606266;
  margin-bottom: 8px;
}

.image-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
  gap: 12px;
  margin-top: 12px;
}

.image-item {
  border-radius: 6px;
  overflow: hidden;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
}

.image-item img {
  width: 100%;
  height: auto;
  display: block;
}

.toolbar-btn {
  padding: 6px 16px;
  border-radius: 4px;
  font-size: 13px;
  cursor: pointer;
  transition: all 0.2s;
  display: inline-flex;
  align-items: center;
  gap: 4px;
  background: #fff;
  border: 1px solid #dcdfe6;
  color: #606266;
}

.toolbar-btn:hover {
  border-color: #409eff;
  color: #409eff;
}

.processing-hint {
  font-size: 12px;
  color: #909399;
  margin-top: 8px;
  text-align: center;
}
</style>

<template>
  <div class="tts-container">
    <div class="main-content">
      <div class="card">
        <div class="card-title">🖼️ 图片生成</div>

        <div class="textarea-wrapper">
          <textarea
              v-model="textInput"
              class="text-input"
              placeholder="请输入图片描述..."
          ></textarea>
        </div>

        <div class="form-row">
          <span class="form-label">历史描述：</span>
          <el-select
              v-model="selectedPrompt"
              placeholder="选择历史描述（可选）"
              filterable
              clearable
              size="small"
              style="width: 300px;"
              @change="onSelectPrompt"
          >
            <el-option
                v-for="prompt in imagePrompts"
                :key="prompt"
                :label="prompt.substring(0, 50) + (prompt.length > 50 ? '...' : '')"
                :value="prompt"
            />
          </el-select>
        </div>

        <div class="form-row">
          <span class="form-label">图片数量：</span>
          <el-input-number
              v-model="imageCount"
              :min="1"
              :max="9"
              :step="1"
              size="small"
              style="width: 100px;"
          />
          <span style="font-size: 12px; color: #909399;">（1-9张）</span>
        </div>

        <div class="form-row">
          <span class="form-label">图片宽度：</span>
          <el-input-number
              v-model="imageWidth"
              :min="512"
              :max="2048"
              :step="8"
              size="small"
              style="width: 120px;"
          />
          <span style="font-size: 12px; color: #909399;">（512-2048，8的倍数）</span>
        </div>

        <div class="form-row">
          <span class="form-label">图片高度：</span>
          <el-input-number
              v-model="imageHeight"
              :min="512"
              :max="2048"
              :step="8"
              size="small"
              style="width: 120px;"
          />
          <span style="font-size: 12px; color: #909399;">（512-2048，8的倍数）</span>
        </div>

        <div class="form-row">
          <span class="form-label">参考图：</span>
          <button class="toolbar-btn" @click="selectReferenceImage">选择参考图</button>
          <span v-if="referenceImageName" style="font-size: 12px; color: #67c23a;">{{ referenceImageName }}</span>
          <button v-if="referenceImageName" class="toolbar-btn" @click="clearReferenceImage" style="margin-left: 8px; padding: 2px 8px;">清除</button>
        </div>

        <div v-if="referenceImageName" class="form-row" style="margin-top: 4px;">
          <span class="form-label"></span>
          <img :src="referenceImagePreview" alt="参考图预览" style="max-width: 200px; max-height: 150px; border-radius: 4px; border: 1px solid #dcdfe6;">
          <span style="font-size: 12px; color: #909399; margin-left: 8px;">参考图将用于图生图</span>
        </div>

        <div style="display: flex; gap: 10px;">
          <button
              class="submit-btn"
              :disabled="!textInput.trim() || generatingImage"
              @click="generateImage"
          >
            {{ generatingImage ? '生成中...' : '🖼️ 生成图片' }}
          </button>
          <button
              class="submit-btn"
              @click="openImageFolder"
              style="background: #409eff;"
          >
            📂 打开文件夹
          </button>
        </div>

        <div v-if="generatingImage" class="processing-hint">
          正在调用 MiniMax API，请稍候...
        </div>

        <div v-if="imageResult" class="result-card" :class="{ error: !imageResult.success }">
          <div v-if="imageResult.success">
            <div class="result-title" style="color: #67c23a;">✅ 图片生成成功</div>
            <div class="result-info">耗时：{{ imageResult.cost }}</div>
            <div class="image-grid">
              <div v-for="(url, index) in imageUrls" :key="index" class="image-item">
                <img :src="url" :alt="'图片 ' + (index + 1)">
              </div>
            </div>
          </div>
          <div v-else>
            <div class="result-title" style="color: #f56c6c;">❌ 图片生成失败</div>
            <div class="result-info">{{ imageResult.message }}</div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import {GenerateImage, SaveImagePrompt, GetImagePrompts, OpenImageFolder, GetImageDataUrl, SelectFile} from "../wailsjs/go/app/App.js";

export default {
  name: 'TextToImage',
  data() {
    return {
      textInput: '',
      imageCount: 9,
      imageWidth: 1920,
      imageHeight: 1080,
      imagePrompts: [],
      selectedPrompt: '',
      imageResult: null,
      generatingImage: false,
      imageUrls: [],
      referenceImageName: '',
      referenceImagePath: '',
      referenceImagePreview: '',
    };
  },
  mounted() {
    this.loadImagePrompts();
  },
  methods: {
    onSelectPrompt(prompt) {
      if (prompt) {
        this.textInput = prompt;
      }
    },

    async loadImagePrompts() {
      try {
        this.imagePrompts = await GetImagePrompts();
      } catch (e) {
        console.error('获取历史描述失败:', e);
      }
    },

    async selectReferenceImage() {
      try {
        const filePath = await SelectFile()
        if (filePath) {
          // 检查文件格式
          const ext = filePath.toLowerCase().split('.').pop()
          if (!['jpg', 'jpeg', 'png'].includes(ext)) {
            this.$message.error('仅支持 JPG、JPEG、PNG 格式')
            return
          }

          this.referenceImageName = filePath.split(/[/\\]/).pop()
          this.referenceImagePath = filePath

          // 预览图片
          this.referenceImagePreview = await GetImageDataUrl(filePath)
        }
      } catch (error) {
        console.error('选择参考图失败:', error)
      }
    },

    clearReferenceImage() {
      this.referenceImageName = ''
      this.referenceImagePath = ''
      this.referenceImagePreview = ''
    },

    getImageUrl(filePath) {
      // 转换文件路径为 URL
      return `file:///${filePath.replace(/\\/g, '/')}`;
    },

    openImageFolder() {
      OpenImageFolder().catch(() => {});
    },

    async generateImage() {
      if (!this.textInput.trim()) {
        this.$message.warning('请输入图片描述文本');
        return;
      }

      this.generatingImage = true;
      this.imageResult = null;

      try {
        // 构建请求参数
        const requestParams = {
          text: this.textInput,
          numImages: this.imageCount,
          width: this.imageWidth,
          height: this.imageHeight,
        };

        // 如果有参考图，传递文件路径
        if (this.referenceImagePath) {
          requestParams.referenceImagePath = this.referenceImagePath;
        }

        console.log('[GenerateImage] 请求参数:', JSON.stringify(requestParams));

        const res = await GenerateImage(requestParams);

        this.imageResult = res;

        if (res.success) {
          this.$message.success('图片生成成功！');
          // 加载图片预览
          this.imageUrls = await Promise.all(res.outputPath.map(path => GetImageDataUrl(path)));
          // 保存描述到数据库
          SaveImagePrompt(this.textInput).then(() => {
            this.loadImagePrompts();
          }).catch(e => {
            console.error('保存描述失败:', e);
          });
        } else {
          this.$message.error('图片生成失败：' + res.message);
        }
      } catch (e) {
        this.$message.error('图片生成失败：' + e);
        this.imageResult = {
          success: false,
          message: e.toString(),
        };
      } finally {
        this.generatingImage = false;
      }
    },
  },
};
</script>
