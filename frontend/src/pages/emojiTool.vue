<style scoped>
.emoji-tabs {
  margin-bottom: 10px;
}

.emoji-grid {
  display: grid;
  grid-template-columns: repeat(20, 1fr);
  gap: 2px;
  max-height: 400px;
  overflow-y: auto;
  padding: 8px;
  border: 1px solid #ddd;
  border-radius: 8px;
  background: #fafafa;
}

.emoji-item {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  font-size: 20px;
  cursor: pointer;
  border-radius: 4px;
  transition: background-color 0.2s;
}

.emoji-item:hover {
  background-color: #e0e0e0;
}

.emoji-item.selected {
  background-color: #409eff;
}

.preview-area {
  display: flex;
  align-items: center;
  gap: 30px;
  margin-top: 20px;
  padding: 20px;
  border: 1px solid #ddd;
  border-radius: 8px;
  background: #fff;
}

.preview-emoji {
  font-size: 100px;
  line-height: 1.2;
}

.preview-right {
  flex: 1;
}

.selected-info {
  font-size: 18px;
  color: #666;
  margin-bottom: 15px;
}
</style>

<template xmlns="http://www.w3.org/1999/html">
  <div style="width: 100%">
    <el-tabs v-model="activeCategory" class="emoji-tabs">
      <el-tab-pane label="😀 表情" name="faces"></el-tab-pane>
      <el-tab-pane label="👋 人物" name="people"></el-tab-pane>
      <el-tab-pane label="🐶 动物" name="animals"></el-tab-pane>
      <el-tab-pane label="🍎 食物" name="food"></el-tab-pane>
      <el-tab-pane label="🚗 旅行" name="travel"></el-tab-pane>
      <el-tab-pane label="💡 物品" name="objects"></el-tab-pane>
      <el-tab-pane label="🔣 符号" name="symbols"></el-tab-pane>
      <el-tab-pane label="🏁 旗帜" name="flags"></el-tab-pane>
    </el-tabs>

    <div class="emoji-grid">
      <div
        v-for="emoji in currentEmojis"
        :key="emoji"
        class="emoji-item"
        :class="{ selected: selectedEmoji === emoji }"
        @click="selectEmoji(emoji)"
        :title="emoji"
      >{{ emoji }}</div>
    </div>

    <div class="preview-area" v-if="selectedEmoji">
      <div class="preview-emoji">{{ selectedEmoji }}</div>
      <div class="preview-right">
        <div style="display: flex; gap: 10px; margin-bottom: 10px;">
          <img v-if="pngUrl" :src="pngUrl" alt="png" style="width: 80px; height: 80px;">
          <a v-if="icoUrl" :href="icoUrl" :download="'emoji_' + selectedEmoji + '.ico'" style="display:inline-block;width:80px;height:80px;line-height:80px;text-align:center;border:1px solid #ddd;border-radius:4px;font-size:12px;color:#666;text-decoration:none;">ICO</a>
        </div>
        <div class="selected-info">已选择: {{ selectedEmoji }}</div>
        <div style="margin-bottom: 10px;">
          <el-button type="primary" @click="copyEmoji">复制Emoji</el-button>
          <el-button type="success" @click="copyToClipboard">复制文字</el-button>
          <el-button type="info" @click="openImageFolder">打开图片文件夹</el-button>
        </div>
              </div>
    </div>
  </div>
</template>

<script>
import {ElNotification} from "element-plus";
import {SaveEmojiToCache, OpenEmojiFolder, GetEmojiImageUrl} from "../wailsjs/go/main/App.js";

export default {
  data() {
    return {
      activeCategory: 'faces',
      selectedEmoji: '',
      pngUrl: '',
      icoUrl: '',
      emojiCategories: {
        faces: ['😀', '😃', '😄', '😁', '😅', '😂', '🤣', '😊', '😇', '🙂', '😉', '😌', '😍', '🥰', '😘', '😋', '😛', '🤪', '🤔', '🤨', '😐', '😑', '😶', '😏', '😒', '🙄', '😬', '🤥', '😔', '😪', '😴', '🤮', '🤧', '😷', '🤒', '🤕', '🤢', '🥵', '🥶', '🥴', '😵', '🤯', '🤠', '🥳', '🥸', '😎', '🤓', '🧐', '😕', '😟', '🙁', '☹️', '😮', '😯', '😲', '😳', '🥺', '😦', '😧', '😨', '😰', '😥', '😢', '😭', '😱', '😖', '😣', '😞', '😓', '😩', '😫', '🥱', '😤', '😡', '😠', '🤬', '😈', '👿', '💀', '☠️', '💩', '🤡', '👹', '👺', '👻', '👽', '👾', '🤖', '😺', '😸', '😹', '😻', '😼', '😽', '🙀', '😿', '😾', '🙈', '🙉', '🙊', '💋', '💌', '💘', '💝', '💟', '❤️', '🧡', '💛', '💚', '💙', '💜', '🖤', '🤍', '🤎', '💔', '❣️', '💕', '💞', '💓', '💗', '💖'],
        people: ['👋', '🤚', '🖐️', '✋', '🖖', '👌', '🤌', '🤏', '✌️', '🤞', '🤟', '🤘', '🤙', '👈', '👉', '👆', '🖕', '👇', '☝️', '👍', '👎', '✊', '👊', '🤛', '🤜', '👏', '🙌', '👐', '🤲', '🤝', '🙏', '✍️', '💅', '🤳', '💪', '🦾', '🦿', '🦵', '🦶', '👂', '🦻', '👃', '🧠', '🫀', '🫁', '🦷', '🦴', '👀', '👁️', '👅', '👄', '👶', '🧒', '👦', '👧', '🧑', '👱', '👨', '🧔', '👩', '🧓', '👴', '👵', '🙍', '🙎', '🙅', '🙆', '💁', '🙋', '🧏', '🙇', '🤦', '🤷', '👮', '🕵️', '💂', '🥷', '👷', '🤴', '👸', '👳', '👲', '🧕', '🤵', '👰', '🤰', '🤱', '👼', '🎒', '👓', '🕶️', '🥽', '🥼', '🧤', '🧣', '🧥', '🧦', '👔', '👕', '👖', '👗', '👘', '👙', '💃', '🩱', '👚', '🪭', '👛', '👜', '👝', '🎅', '🎄', '🧝', '🧙', '🧚', '🧛', '🧜', '🧞', '🏊', '🤽', '🚣', '🧗', '🚵', '🏋️', '🤼', '🤸', '⛹️', '🤺', '🤾', '🏌️', '🏇', '🧘', '🏄', '🛀'],
        animals: ['🐶', '🐱', '🐭', '🐹', '🐰', '🦊', '🐻', '🐼', '🐻‍❄️', '🐨', '🐯', '🦁', '🐮', '🐷', '🐽', '🐸', '🐵', '🙈', '🙉', '🙊', '🐒', '🐔', '🐧', '🐦', '🐤', '🐣', '🐥', '🦆', '🦅', '🦉', '🦇', '🐺', '🐗', '🐴', '🦄', '🐝', '🪱', '🐛', '🦋', '🐌', '🐞', '🐜', '🪰', '🪲', '🪳', '🦟', '🦗', '🕷️', '🕸️', '🦂', '🐢', '🐍', '🦎', '🦖', '🦕', '🐙', '🦑', '🦐', '🦞', '🦀', '🐡', '🐠', '🐟', '🐬', '🐳', '🐋', '🦈', '🐊', '🐅', '🐆', '🦓', '🦍', '🦧', '🦣', '🐘', '🦛', '🦏', '🐪', '🐫', '🦒', '🦘', '🦬', '🐃', '🐂', '🐄', '🐎', '🐖', '🐏', '🐑', '🦙', '🐐', '🦌', '🐕', '🐩', '🦮', '🐈', '🐈‍⬛', '🐓', '🦃', '🦚', '🦜', '🦢', '🦩', '🕊️', '🐇', '🦝', '🦨', '🦡', '🦫', '🦦', '🦥', '🐁', '🐀', '🐿️', '🦔', '🌸', '🌺', '🌻', '🌹', '🥀', '🌷', '💐', '🌼', '🍀', '🍁', '🍂', '🍃', '🌿', '🌲', '🌳', '🌴', '🌵', '🎄', '🌲', '🌳', '🌴', '🌵', '🎄', '☘️', '🍀', '🍁', '🍂', '🍃', '🪴', '🎍', '🎋', '🌾', '🌱', '🌪️', '🪨', '☀️', '🌙', '🌝', '🌞', '⭐', '🌟', '💫', '🌈', '☁️', '⛅', '🌤️', '☀️', '🌦️', '🌧️', '⛈️', '🌩️', '🌨️', '☀️', '🌤️', '⛅', '🌥️', '☁️', '🌧️', '⛈️', '🌩️', '🌨️', '❄️', '☃️', '⛄', '🌬️', '💨', '🌪️', '🌫️', '🌪️'],
        food: ['🍎', '🍐', '🍊', '🍋', '🍌', '🍉', '🍇', '🍓', '🫐', '🍈', '🍒', '🍑', '🥭', '🍍', '🥥', '🥝', '🍅', '🍆', '🥑', '🥦', '🥬', '🥒', '🌶️', '🫑', '🌽', '🥕', '🫒', '🧄', '🧅', '🥔', '🍠', '🥐', '🥖', '🍞', '🥨', '🥯', '🧇', '🥞', '🧈', '🍳', '🥚', '🧀', '🥓', '🥩', '🍗', '🍖', '🌭', '🍔', '🍟', '🍕', '🫓', '🥪', '🥙', '🧆', '🌮', '🌯', '🫔', '🥗', '🥘', '🫕', '🍝', '🍜', '🍲', '🍛', '🍣', '🍱', '🥟', '🦪', '🍤', '🍙', '🍚', '🍘', '🍥', '🥠', '🥮', '🍢', '🍡', '🍧', '🍨', '🍦', '🥧', '🧁', '🍰', '🎂', '🍮', '🍭', '🍬', '🍫', '🍿', '🍩', '🍪', '🌰', '🥜', '🍯', '🧃', '🥤', '🧋', '🍵', '☕', '🫖', '🥛', '🍼', '🍺', '🍻', '🍷', '🥂', '🥃', '🍸', '🍹', '🧉', '🍾', '🧊', '🥄', '🍴', '🍽️', '🥣', '🥡', '🥢', '🧂'],
        travel: ['🚗', '🚕', '🚙', '🚌', '🚎', '🏎️', '🚓', '🚑', '🚒', '🚐', '🛻', '🚚', '🚛', '🚜', '🏍️', '🛵', '🚲', '🛴', '🛺', '🚨', '🚔', '🚍', '🚘', '🚖', '🚡', '🚠', '🚟', '🚃', '🚋', '🚞', '🚝', '🚄', '🚅', '🚈', '🚂', '🚆', '🚇', '🚊', '🚉', '✈️', '🛫', '🛬', '🛩️', '💺', '🛰️', '🚀', '🛸', '🚁', '🛶', '⛵', '🚤', '🛥️', '🛳️', '⛴️', '🚢', '⚓', '🪝', '⛽', '🚧', '🚦', '🚥', '🚏', '🗺️', '🗿', '🗽', '🗼', '🏰', '🏯', '🏟️', '🎡', '🎢', '🎠', '⛲', '⛱️', '🏖️', '🏝️', '🏜️', '🌋', '⛰️', '🏔️', '🗻', '🏕️', '⛺', '🛖', '🏠', '🏡', '🏘️', '🏚️', '🏗️', '🏭', '🏢', '🏬', '🏣', '🏤', '🏥', '🏦', '🏨', '🏪', '🏫', '🏩', '💒', '🏛️', '⛪', '🕌', '🕍', '🛕', '🕋', '⛩️', '🛤️', '🛣️', '🗾', '🎑', '🏞️', '🎐', '🎃', '🎆', '🎇', '🧨', '✨', '🎈', '🎉', '🎊', '🎋', '🎍', '🎎', '🧧', '🎀', '🎁', '🎗️', '🎟️', '🎫', '🎖️', '🏆', '🏅', '🥇', '🥈', '🥉', '🏵️', '🎪', '🤹', '🎭', '🃏', '🀄', '🎴', '🎱', '🎯', '🎳', '🎮', '🎰', '🧩', '🪅', '🪆', '🪄', '🎲', '♟️', '🎺', '🎷', '🎸', '🎹', '🎚️', '🎛️', '🎙️', '🎤', '🎧', '📻', '🎬', '🎥', '📽️', '🎞️', '📷', '📸', '📹', '💡', '🔦', '🏮', '🪔'],
        objects: ['📱', '📲', '💻', '🖥️', '🖨️', '⌨️', '🖱️', '🖲️', '💽', '💾', '💿', '📀', '📼', '📞', '☎️', '📟', '📠', '📺', '📡', '🔋', '🔌', '💵', '💴', '💶', '💷', '💰', '💳', '💎', '⚖️', '🪜', '🧰', '🪛', '🔧', '🔨', '⚒️', '🛠️', '⛏️', '🪚', '🔩', '⚙️', '🪤', '🧱', '⛓️', '🧲', '🔫', '💣', '🪓', '🔪', '🗡️', '⚔️', '🛡️', '🚬', '⚰️', '🪦', '⚱️', '🏺', '🔮', '📿', '🧿', '💈', '⚗️', '🔭', '🔬', '🕳️', '🩹', '🩺', '💊', '💉', '🩸', '🧬', '🦠', '🧫', '🧪', '🌡️', '🧹', '🪠', '🧺', '🧻', '🚽', '🚰', '🚿', '🛁', '🧼', '🪥', '🪒', '🧽', '🪣', '🧴', '🛒', '📌', '📍', '✂️', '🖊️', '🖋️', '✒️', '🖌️', '🖍️', '📝', '📁', '📂', '🗂️', '📅', '📆', '🗒️', '🗓️', '📇', '📈', '📉', '📊', '📋', '📎', '🖇️', '📏', '📐', '🗃️', '🗄️', '🗑️', '🔍', '🔎', '🔏', '🔐', '🔒', '🔓', '🗝️', '🔑', '🕰️', '⌛', '⏳', '⌚', '⏰', '⏱️', '⏲️', '🕐', '🕑', '🕒', '🕓', '🕔', '🕕', '🕖', '🕗', '🕘', '🕙', '🕚', '🕛'],
        symbols: ['❤️', '🧡', '💛', '💚', '💙', '💜', '🖤', '🤍', '🤎', '💔', '❣️', '💕', '💞', '💓', '💗', '💖', '☮️', '✝️', '☯️', '🕉️', '☪️', '🔯', '🛐', '⛎', '♈', '♉', '♊', '♋', '♌', '♍', '♎', '♏', '♐', '♑', '♒', '♓', '🆔', '⚛️', '🉑', '☢️', '☣️', '📴', '📳', '🈶', '🈚', '🈸', '🈺', '🈷️', '✴️', '🆚', '💮', '🉐', '㊙️', '㊗️', '🈴', '🈵', '🈹', '🈲', '🅰️', '🅱️', '🆎', '🆑', '🅾️', '🆘', '❌', '⭕', '🛑', '⛔', '📛', '🚫', '💯', '💢', '♨️', '🚷', '🚯', '🚳', '🚱', '🔞', '📵', '🚭', '❗', '❕', '❓', '❔', '‼️', '⁉️', '🔅', '🔆', '〽️', '⚠️', '🚸', '🔱', '⚜️', '🔰', '♻️', '✅', '🈯', '💹', '❇️', '✳️', '❎', '🌐', '💠', 'Ⓜ️', '🌀', '💤', '🏧', '🚾', '♿', '🅿️', '🛗', '🈳', '🈂️', '🛂', '🛃', '🛄', '🛅', '🚹', '🚺', '🚼', '⚧️', '🚻', '🚮', '🎦', '📶', '🈁', '🔣', 'ℹ️', '🔤', '🔡', '🔠', '🆖', '🆗', '🆙', '🆒', '🆕', '🆓', '0️⃣', '1️⃣', '2️⃣', '3️⃣', '4️⃣', '5️⃣', '6️⃣', '7️⃣', '8️⃣', '9️⃣', '🔟', '🔢', '#️⃣', '*️⃣', '⏏️', '▶️', '⏸️', '⏯️', '⏹️', '⏺️', '⏭️', '⏮️', '⏩', '⏪', '⏫', '⏬', '◀️', '🔼', '🔽', '➡️', '⬅️', '⬆️', '⬇️', '↗️', '↘️', '↙️', '↖️', '↕️', '↔️', '↪️', '↩️', '⤴️', '⤵️', '🔀', '🔁', '🔂', '🔄', '🔃', '🎵', '🎶', '➕', '➖', '➗', '✖️', '♾️', '💲', '💱', '™️', '©️', '®️', '♠️', '♥️', '♣️', '♦️', '🏁', '🚩', '🎌', '🏴', '🏳️', '🏳️‍🌈', '🏳️‍⚧️', '🏴‍☠️', '☂️', '🧳', '🎏', '🔔', '🔕', '📢', '📣', '💬', '💭', '🗨️', '🗯️', '🔃', '🔄', '🔙', '🔚', '🔛', '🔜', '🔝', '☑️', '🔘', '🔴', '🟠', '🟡', '🟢', '🔵', '🟣', '⚫', '⚪', '🟤', '🔺', '🔻', '🔸', '🔹', '💠', '🔷', '🔶', '💢', '💨', '💦', '👁️', '🦰', '🦱', '🦳', '🦲'],
        flags: ['🏴󠁧󠁢󠁥󠁮󠁧󠁿', '🏴󠁧󠁢󠁳󠁣󠁴󠁿', '🏴󠁧󠁢󠁷󠁣󠁴󠁿']
      }
    }
  },
  computed: {
    currentEmojis() {
      return this.emojiCategories[this.activeCategory] || []
    }
  },
  methods: {
    selectEmoji(emoji) {
      this.selectedEmoji = emoji;
      this.generateImage();
    },
    async generateImage() {
      if (!this.selectedEmoji) return;
      try {
        const id = await SaveEmojiToCache(this.selectedEmoji);
        if (!id) {
          this.$message.error("图片生成失败: 无法获取ID");
          return;
        }
        this.pngUrl = await GetEmojiImageUrl(this.selectedEmoji, "png");
        this.icoUrl = await GetEmojiImageUrl(this.selectedEmoji, "ico");
      } catch (err) {
        this.$message.error("图片生成异常:" + err);
      }
    },
    copyEmoji() {
      navigator.clipboard.writeText(this.selectedEmoji).then(() => {
        ElNotification({
          title: '复制成功',
          message: 'Emoji已复制到剪贴板',
          position: 'bottom-right',
          type: 'success',
        });
      }).catch(() => {
        this.$message.error('复制失败');
      });
    },
    copyToClipboard() {
      navigator.clipboard.writeText(this.selectedEmoji).then(() => {
        ElNotification({
          title: '复制成功',
          message: '文字已复制到剪贴板',
          position: 'bottom-right',
          type: 'success',
        });
      }).catch(() => {
        this.$message.error('复制失败');
      });
    },
    openImageFolder() {
      OpenEmojiFolder().catch(err => {
        this.$message.error("打开文件夹失败: " + err);
      });
    }
  }
}
</script>
