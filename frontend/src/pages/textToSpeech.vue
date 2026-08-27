<style scoped>
.tts-container {
  width: 100%;
  height: 100%;
  padding: 0;
  box-sizing: border-box;
  background: #f5f7fa;
  display: flex;
  flex-direction: column;
}

.main-content {
  width: 100%;
}

.card {
  background: #fff;
  border-radius: 8px;
  padding: 20px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.08);
  margin-bottom: 20px;
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
  min-height: 80px;
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

.char-count {
  font-size: 12px;
  color: #909399;
  text-align: right;
  margin-top: 4px;
}

.submit-btn {
  width: 100%;
  padding: 10px 20px;
  background: #409eff;
  border: none;
  border-radius: 4px;
  color: #fff;
  font-size: 14px;
  cursor: pointer;
  transition: background 0.2s;
}

.submit-btn:hover {
  background: #66b1ff;
}

.submit-btn:disabled {
  background: #c0c4cc;
  cursor: not-allowed;
}

.submit-btn.clone-btn {
  background: #e6a23c;
}

.submit-btn.clone-btn:hover {
  background: #ebb563;
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

.result-actions {
  display: flex;
  gap: 10px;
  margin-top: 12px;
}

.audio-player {
  width: 100%;
  margin-top: 12px;
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
}

.toolbar-btn.primary {
  background: #409eff;
  border: none;
  color: #fff;
}

.toolbar-btn.primary:hover {
  background: #66b1ff;
}

.toolbar-btn {
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

.upload-area {
  border: 2px dashed #dcdfe6;
  border-radius: 6px;
  padding: 20px;
  text-align: center;
  cursor: pointer;
  transition: border-color 0.2s;
}

.upload-area:hover {
  border-color: #409eff;
}

.upload-area.dragover {
  border-color: #409eff;
  background: #f0f9eb;
}

.voice-id-input {
  width: 200px;
  padding: 8px 12px;
  border: 1px solid #dcdfe6;
  border-radius: 4px;
  font-size: 13px;
}

.voice-id-input:focus {
  outline: none;
  border-color: #409eff;
}

.custom-voice-tag {
  display: inline-block;
  padding: 4px 8px;
  background: #fdf6ec;
  border: 1px solid #f5dab1;
  border-radius: 4px;
  font-size: 12px;
  color: #e6a23c;
  margin: 2px;
  cursor: pointer;
}

.custom-voice-tag:hover {
  background: #fdf6ec;
  border-color: #e6a23c;
}
</style>

<template>
  <div class="tts-container">
    <div class="main-content">
      <el-tabs v-model="activeTab" type="border-card">
        <!-- 语音合成 -->
        <el-tab-pane label="🔊 文字转语音" name="tts">
          <div class="card">
            <div class="textarea-wrapper">
              <textarea
                  v-model="textInput"
                  class="text-input"
                  placeholder="请输入要转换的文本..."
                  @input="updateCharCount"
              ></textarea>
              <div class="char-count">字数：{{ charCount }}</div>
            </div>

            <div class="form-row">
              <span class="form-label">音色：</span>
              <el-select
                  v-model="selectedVoiceId"
                  placeholder="选择音色"
                  size="small"
                  style="width: 200px;"
                  @change="onVoiceChange"
              >
                <el-option-group
                    v-for="group in voiceGroups"
                    :key="group.label"
                    :label="group.label"
                >
                  <el-option
                      v-for="voice in group.options"
                      :key="voice.id"
                      :label="voice.name"
                      :value="voice.id"
                  />
                </el-option-group>
              </el-select>
              <span style="font-size: 12px; color: #909399;">({{ selectedVoice.name }})</span>
            </div>

            <!-- 自定义音色快捷选择 -->
            <div v-if="customVoices.length > 0" class="form-row">
              <span class="form-label">我的音色：</span>
              <span
                  v-for="voiceId in customVoices"
                  :key="voiceId"
                  class="custom-voice-tag"
                  @click="selectCustomVoice(voiceId)"
              >
                {{ voiceId }}
              </span>
            </div>

            <button
                class="submit-btn"
                :disabled="!textInput.trim() || converting"
                @click="convertToSpeech"
            >
              {{ converting ? '生成中...' : '🎙️ 生成语音' }}
            </button>

            <div v-if="converting" class="processing-hint">
              正在调用 MiniMax API，请稍候...
            </div>

            <div v-if="ttsResult" class="result-card" :class="{ error: !ttsResult.success }">
              <div v-if="ttsResult.success">
                <div class="result-title" style="color: #67c23a;">✅ 语音生成成功</div>
                <div class="result-info">耗时：{{ ttsResult.cost }}</div>
                <audio v-if="audioSrc" controls class="audio-player">
                  <source :src="audioSrc" type="audio/mpeg">
                  您的浏览器不支持音频播放
                </audio>
                <div class="result-actions">
                  <button class="toolbar-btn primary" @click="openAudioFolder">📂 打开文件夹</button>
                </div>
              </div>
              <div v-else>
                <div class="result-title" style="color: #f56c6c;">❌ 语音生成失败</div>
                <div class="result-info">{{ ttsResult.message }}</div>
              </div>
            </div>
          </div>
        </el-tab-pane>

        <!-- 音色克隆 -->
        <el-tab-pane label="🎭 音色复刻" name="clone">
          <div class="card">
            <div class="form-row">
              <span class="form-label">自定义音色ID：</span>
              <input
                  v-model="cloneVoiceId"
                  class="voice-id-input"
                  placeholder="输入自定义音色ID"
              />
              <span style="font-size: 12px; color: #909399;">(英文、数字、下划线)</span>
            </div>

            <div class="form-row">
              <span class="form-label">上传音频：</span>
              <div
                  class="upload-area"
                  :class="{ dragover: isDragOver }"
                  @click="triggerFileInput"
                  @dragover.prevent="isDragOver = true"
                  @dragleave="isDragOver = false"
                  @drop.prevent="handleFileDrop"
              >
                <input
                    type="file"
                    ref="fileInput"
                    accept=".mp3,.m4a,.wav"
                    style="display: none"
                    @change="handleFileSelect"
                />
                <div v-if="!selectedFile">
                  <div style="font-size: 24px;">📤</div>
                  <div style="color: #606266;">点击或拖拽上传音频文件</div>
                  <div style="font-size: 12px; color: #909399;">支持 mp3、m4a、wav 格式，时长 10秒-5分钟，大小不超过 20MB</div>
                </div>
                <div v-else>
                  <div style="font-size: 24px;">✅</div>
                  <div style="color: #67c23a;">{{ selectedFile.name }}</div>
                  <div style="font-size: 12px; color: #909399;">{{ (selectedFile.size / 1024 / 1024).toFixed(2) }} MB</div>
                </div>
              </div>
            </div>

            <div class="form-row">
              <span class="form-label">参考文本：</span>
              <input
                  v-model="referenceText"
                  class="voice-id-input"
                  placeholder="输入参考文本（可选，用于增强克隆效果）"
                  style="width: 400px;"
              />
            </div>

            <button
                class="submit-btn clone-btn"
                :disabled="!cloneVoiceId.trim() || !selectedFile || cloning"
                @click="cloneVoice"
            >
              {{ cloning ? '复刻中...' : '🎭 复刻音色' }}
            </button>

            <div v-if="cloning" class="processing-hint">
              正在复刻音色，请稍候...
            </div>

            <div v-if="cloneResult" class="result-card" :class="{ error: !cloneResult.success }">
              <div v-if="cloneResult.success">
                <div class="result-title" style="color: #67c23a;">✅ {{ cloneResult.message }}</div>
              </div>
              <div v-else>
                <div class="result-title" style="color: #f56c6c;">❌ 复刻失败</div>
                <div class="result-info">{{ cloneResult.message }}</div>
              </div>
            </div>
          </div>
        </el-tab-pane>
      </el-tabs>
    </div>
  </div>
</template>

<script>
import {TextToSpeech, GetAvailableVoices, OpenAudioFolder, GetAudioDataUrl, CloneVoice, ListCustomVoices} from "../wailsjs/go/app/App.js";

export default {
  name: 'TextToSpeech',
  data() {
    return {
      activeTab: 'tts',
      textInput: '',
      charCount: 0,
      voices: [],
      selectedVoice: { id: 'male-qn-qingse', name: '青涩青年音色', language: '中文(普通话)' },
      selectedVoiceId: 'male-qn-qingse',
      converting: false,
      ttsResult: null,
      audioSrc: '',
      // 音色克隆
      cloneVoiceId: '',
      selectedFile: null,
      referenceText: '',
      cloning: false,
      cloneResult: null,
      customVoices: [],
      isDragOver: false,
    };
  },
  mounted() {
    this.loadVoices();
    this.loadCustomVoices();
  },
  computed: {
    voiceGroups() {
      const groups = {};
      this.voices.forEach(voice => {
        const lang = voice.language || '其他';
        if (!groups[lang]) {
          groups[lang] = [];
        }
        groups[lang].push(voice);
      });
      return Object.keys(groups).map(lang => ({
        label: lang,
        options: groups[lang]
      }));
    }
  },
  methods: {
    updateCharCount() {
      this.charCount = this.textInput.length;
    },

    onVoiceChange(voiceId) {
      const voice = this.voices.find(v => v.id === voiceId);
      if (voice) {
        this.selectedVoice = voice;
      }
    },

    selectCustomVoice(voiceId) {
      this.selectedVoiceId = voiceId;
      this.selectedVoice = { id: voiceId, name: voiceId, language: '自定义' };
    },

    async loadVoices() {
      try {
        this.voices = await GetAvailableVoices();
        if (this.voices.length > 0) {
          this.selectedVoice = this.voices[0];
          this.selectedVoiceId = this.voices[0].id;
        }
      } catch (e) {
        console.error('获取音色列表失败:', e);
        this.voices = [
          { id: 'male-qn-qingse', name: '青涩青年音色', language: '中文(普通话)' },
          { id: 'female-shaonv', name: '少女音色', language: '中文(普通话)' },
        ];
        this.selectedVoice = this.voices[0];
      }
    },

    async loadCustomVoices() {
      try {
        const res = await ListCustomVoices();
        if (res.success) {
          this.customVoices = res.custom_voices || [];
        }
      } catch (e) {
        console.error('获取自定义音色失败:', e);
      }
    },

    async convertToSpeech() {
      if (!this.textInput.trim()) {
        this.$message.warning('请输入要转换的文本');
        return;
      }

      this.converting = true;
      this.ttsResult = null;
      this.audioSrc = '';

      try {
        const res = await TextToSpeech({
          text: this.textInput,
          voiceId: this.selectedVoice.id,
        });

        this.ttsResult = res;

        if (res.success) {
          this.$message.success('语音生成成功！');
          GetAudioDataUrl(res.outputPath).then(dataUrl => {
            this.audioSrc = dataUrl;
          });
        } else {
          this.$message.error('语音生成失败：' + res.message);
        }
      } catch (e) {
        this.$message.error('语音生成失败：' + e);
        this.ttsResult = {
          success: false,
          message: e.toString(),
        };
      } finally {
        this.converting = false;
      }
    },

    openAudioFolder() {
      OpenAudioFolder();
    },

    triggerFileInput() {
      this.$refs.fileInput.click();
    },

    handleFileSelect(e) {
      const file = e.target.files[0];
      if (file) {
        this.validateAndSetFile(file);
      }
    },

    handleFileDrop(e) {
      this.isDragOver = false;
      const file = e.dataTransfer.files[0];
      if (file) {
        this.validateAndSetFile(file);
      }
    },

    validateAndSetFile(file) {
      const validExts = ['.mp3', '.m4a', '.wav'];
      const ext = '.' + file.name.split('.').pop().toLowerCase();

      if (!validExts.includes(ext)) {
        this.$message.error('不支持的文件格式，请上传 mp3、m4a、wav 格式');
        return;
      }

      if (file.size > 20 * 1024 * 1024) {
        this.$message.error('文件大小超过 20MB');
        return;
      }

      this.selectedFile = file;
    },

    async cloneVoice() {
      if (!this.cloneVoiceId.trim()) {
        this.$message.warning('请输入自定义音色ID');
        return;
      }

      if (!this.selectedFile) {
        this.$message.warning('请上传待克隆音频');
        return;
      }

      this.cloning = true;
      this.cloneResult = null;

      try {
        // 注意：这里需要先上传文件获取 file_id
        // 简化处理：直接调用克隆接口（实际应先上传文件）
        const res = await CloneVoice({
          source_file_id: this.selectedFile.name, // 实际应该先上传获取真实file_id
          reference_text: this.referenceText,
          voice_id: this.cloneVoiceId,
        });

        this.cloneResult = res;

        if (res.success) {
          this.$message.success('音色克隆成功！');
          this.loadCustomVoices();
          // 清空表单
          this.selectedFile = null;
          this.referenceText = '';
        } else {
          this.$message.error('音色克隆失败：' + res.message);
        }
      } catch (e) {
        this.$message.error('音色克隆失败：' + e);
        this.cloneResult = {
          success: false,
          message: e.toString(),
        };
      } finally {
        this.cloning = false;
      }
    },
  },
};
</script>