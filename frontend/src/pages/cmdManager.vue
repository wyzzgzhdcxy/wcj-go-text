<style scoped>
.cmd-manager {
  width: 100%;
  height: 100%;
  display: flex;
  flex-direction: column;
  box-sizing: border-box;
}

.toolbar {
  margin-bottom: 12px;
  flex-shrink: 0;
}

.cmd-table {
  flex: 1;
  width: 100%;
}

.command-input {
  font-family: Consolas, Monaco, monospace;
}

.output-dialog {
  font-family: Consolas, Monaco, monospace;
  white-space: pre-wrap;
  word-break: break-all;
}
</style>

<template>
  <div class="cmd-manager">
    <div class="toolbar">
      <el-button type="primary" @click="showAddDialog">添加命令行</el-button>
    </div>

    <el-table :data="commands" border class="cmd-table" v-loading="loading">
      <el-table-column prop="id" label="ID" width="80"/>
      <el-table-column prop="name" label="名称" width="150"/>
      <el-table-column prop="command" label="命令" min-width="300">
        <template #default="scope">
          <code class="command-input">{{ scope.row.command }}</code>
        </template>
      </el-table-column>
      <el-table-column prop="description" label="描述" min-width="200"/>
      <el-table-column prop="updated_at" label="更新时间" width="180"/>
      <el-table-column label="操作" width="220" fixed="right">
        <template #default="scope">
          <el-button size="small" type="success" @click="executeCommand(scope.row)">执行</el-button>
          <el-button size="small" type="primary" @click="editCommand(scope.row)">编辑</el-button>
          <el-button size="small" type="danger" @click="deleteCommand(scope.row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <!-- 添加/编辑对话框 -->
    <el-dialog v-model="dialogVisible" :title="dialogTitle" width="600px">
      <el-form :model="form" label-width="80px">
        <el-form-item label="名称">
          <el-input v-model="form.name" placeholder="请输入命令名称"/>
        </el-form-item>
        <el-form-item label="命令">
          <el-input v-model="form.command" type="textarea" :rows="3" placeholder="请输入命令行"/>
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="form.description" type="textarea" :rows="2" placeholder="请输入描述"/>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="saveCommand">确定</el-button>
      </template>
    </el-dialog>

    <!-- 执行结果对话框 -->
    <el-dialog v-model="outputVisible" title="执行结果" width="800px">
      <el-input v-model="commandOutput" type="textarea" :rows="15" readonly class="output-dialog"/>
      <template #footer>
        <el-button @click="outputVisible = false">关闭</el-button>
      </template>
    </el-dialog>

    <!-- 占位符参数输入对话框 -->
    <el-dialog v-model="paramVisible" title="输入参数" width="500px">
      <el-form :model="paramForm" label-width="100px">
        <el-form-item v-for="(placeholder, index) in placeholders" :key="index" :label="placeholder.label">
          <el-input v-model="paramForm.values[placeholder.index]" :placeholder="placeholder.desc"/>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="paramVisible = false">取消</el-button>
        <el-button type="primary" @click="executeWithParams">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script>
import {
  ListCmdCommands,
  AddCmdCommand,
  UpdateCmdCommand,
  DeleteCmdCommand,
  ExecuteCmdCommand
} from "../wailsjs/go/app/App.js";
import {ElMessage} from "element-plus";

export default {
  data() {
    return {
      commands: [],
      loading: false,
      dialogVisible: false,
      outputVisible: false,
      paramVisible: false,
      dialogTitle: "添加命令行",
      commandOutput: "",
      form: {
        id: 0,
        name: "",
        command: "",
        description: ""
      },
      placeholders: [],
      paramForm: {
        command: "",
        values: []
      }
    };
  },

  mounted() {
    this.loadCommands();
  },

  methods: {
    loadCommands() {
      this.loading = true;
      ListCmdCommands().then(result => {
        this.commands = result || [];
        this.loading = false;
      }).catch(err => {
        ElMessage.error("加载命令行失败: " + err);
        this.loading = false;
      });
    },

    showAddDialog() {
      this.dialogTitle = "添加命令行";
      this.form = {
        id: 0,
        name: "",
        command: "",
        description: ""
      };
      this.dialogVisible = true;
    },

    editCommand(row) {
      this.dialogTitle = "编辑命令行";
      this.form = {
        id: row.id,
        name: row.name,
        command: row.command,
        description: row.description || ""
      };
      this.dialogVisible = true;
    },

    saveCommand() {
      if (!this.form.name || !this.form.command) {
        ElMessage.warning("名称和命令不能为空");
        return;
      }

      if (this.form.id === 0) {
        // 添加
        AddCmdCommand(this.form).then(id => {
          ElMessage.success("添加成功");
          this.dialogVisible = false;
          this.loadCommands();
        }).catch(err => {
          ElMessage.error("添加失败: " + err);
        });
      } else {
        // 更新
        UpdateCmdCommand(this.form).then(() => {
          ElMessage.success("更新成功");
          this.dialogVisible = false;
          this.loadCommands();
        }).catch(err => {
          ElMessage.error("更新失败: " + err);
        });
      }
    },

    deleteCommand(row) {
      this.$confirm(`确定要删除命令 "${row.name}" 吗?`, "提示", {
        confirmButtonText: "确定",
        cancelButtonText: "取消",
        type: "warning"
      }).then(() => {
        DeleteCmdCommand(row.id).then(() => {
          ElMessage.success("删除成功");
          this.loadCommands();
        }).catch(err => {
          ElMessage.error("删除失败: " + err);
        });
      }).catch(() => {
      });
    },

    executeCommand(row) {
      // 检测命令中的占位符
      const placeholders = this.detectPlaceholders(row.command, row.description);
      if (placeholders.length > 0) {
        // 有占位符，弹出参数输入框
        this.placeholders = placeholders;
        this.paramForm = {
          command: row.command,
          id: row.id,
          values: new Array(placeholders.length).fill("")
        };
        this.paramVisible = true;
      } else {
        // 无占位符，直接执行
        this.doExecute(row.id, row.command);
      }
    },

    detectPlaceholders(command, description) {
      // 匹配 %d(数字), %s(字符串), %f(浮点数), %(文件名) 等占位符
      const regex = /%[dsf]|\%\{[^}]+\}/g;
      const matches = command.match(regex) || [];

      // 取描述的第一行，然后按逗号分割
      const firstLine = description ? description.split("\n")[0].trim() : "";
      const descParts = firstLine ? firstLine.split(",").map(s => s.trim()) : [];

      const placeholders = [];
      matches.forEach((match, index) => {
        let label = match;
        let desc = "";
        const pattern = match; // 保存原始占位符模式用于替换

        // 使用描述分割后的对应部分作为标签
        if (descParts[index]) {
          label = descParts[index];
          desc = `请输入 ${label}`;
        } else {
          // 没有描述时使用默认值
          if (match === "%d") {
            desc = "请输入数字";
          } else if (match === "%s") {
            desc = "请输入字符串";
          } else if (match === "%f") {
            desc = "请输入浮点数";
          } else if (match.startsWith("%{")) {
            desc = `请输入 ${match.slice(2, -1)}`;
          }
        }
        placeholders.push({ pattern, label, desc, index });
      });
      return placeholders;
    },

    executeWithParams() {
      // 检查是否所有参数都已输入
      const hasEmpty = this.paramForm.values.some(v => v === "" || v === undefined);
      if (hasEmpty) {
        ElMessage.warning("请填写所有参数");
        return;
      }

      let command = this.paramForm.command;
      let paramIndex = 0; // 参数值的索引

      // 逐个替换占位符
      const regex = /%[dsf]|\%\{[^}]+\}/g;
      let match;
      while ((match = regex.exec(command)) !== null) {
        if (paramIndex >= this.paramForm.values.length) break;
        const value = this.paramForm.values[paramIndex];
        const pos = match.index;
        command = command.substring(0, pos) + value + command.substring(pos + match[0].length);
        paramIndex++;
      }

      this.paramVisible = false;
      this.doExecute(this.paramForm.id, command);
    },

    doExecute(id, command) {
      // 如果 command 与数据库中的原始命令不同，说明有占位符被替换
      const originalCommand = this.commands.find(c => c.id === id)?.command;
      const hasReplace = originalCommand !== command;

      // 始终传递命令字符串，由后端决定是否使用传入的值
      ExecuteCmdCommand(id, hasReplace ? command : originalCommand).then(output => {
        this.commandOutput = "命令: " + command + "\n\n" + (output || "(无输出)");
        this.outputVisible = true;
      }).catch(err => {
        this.commandOutput = "命令: " + command + "\n\n执行失败: " + err;
        this.outputVisible = true;
      });
    }
  }
};
</script>
