<template>
  <div class="card-div" style="font-size: 14px; color: #99a9bf">
    <el-form :model="podSpec" :rules="rules" label-width="120px"
      label-position="left" size="small">
      <el-form-item label="指定节点" prop="nodeName">
        <el-select v-model="podSpec.nodeName" size="small" class="input-class">
          <el-option label="不指定" value=""></el-option>
          <el-option :label="n.name" :value="n.name" :key="i" v-for="(n, i) in nodes"></el-option>
        </el-select>
      </el-form-item>
      <el-form-item label="指定调度器" prop="schedulerName">
        <el-input v-model="podSpec.schedulerName" size="small" class="input-class" placeholder="指定调度器名称"></el-input>
      </el-form-item>
      <el-form-item label="选择节点标签" prop="nodeSelector">
        <el-row style="margin-bottom: -15px;" v-if="podSpec.nodeSelector.length > 0">
          <el-col :span="7">
            <div class="border-span-header">
              <span  class="border-span-content">*</span>标签键
            </div>
          </el-col>
          <el-col :span="17">
            <div class="border-span-header">
              <span  class="border-span-content">*</span>标签值
            </div>
          </el-col>
        </el-row>
        <el-row style="padding-top: 10px;" v-for="(item, idx) in podSpec.nodeSelector" :key="idx">
          <el-col :span="7" style="padding-right: 10px;">
            <div class="border-span-header">
              <el-input v-model="item.key" size="small" placeholder="节点标签键"></el-input>
            </div>
          </el-col>
          <el-col :span="7" style="padding-right: 10px;">
            <div class="border-span-header">
              <el-input v-model="item.value" size="small" placeholder="节点标签值"></el-input>
            </div>
          </el-col>
          <el-col :span="3">
            <el-button plain size="mini" style="padding-left: 10px; padding-right: 10px;" 
              @click="podSpec.nodeSelector.splice(idx, 1)" icon="el-icon-minus"></el-button>
          </el-col>
        </el-row>
        <el-button plain size="mini" @click="podSpec.nodeSelector.push({})" icon="el-icon-plus"></el-button>
      </el-form-item>
      <el-form-item label="污点容忍" prop="toleration">
        <el-row style="margin-bottom: -15px;" v-if="podSpec.tolerations.length > 0">
          <el-col :span="4">
            <div class="border-span-header">
              标签键
            </div>
          </el-col>
          <el-col :span="3">
            <div class="border-span-header">
              运算符
            </div>
          </el-col>
          <el-col :span="4">
            <div class="border-span-header">
              标签值
            </div>
          </el-col>
          <el-col :span="4">
            <div class="border-span-header">
              影响
            </div>
          </el-col>
          <el-col :span="8">
            <div class="border-span-header">
              时间
            </div>
          </el-col>
        </el-row>
        <el-row style="padding-top: 10px;" v-for="(item, idx) in podSpec.tolerations" :key="idx">
          <el-col :span="4" style="padding-right: 10px;">
            <div class="item-content">
              <el-input v-model="item.key" size="small" placeholder=""></el-input>
            </div>
          </el-col>
          <el-col :span="3" style="padding-right: 10px;">
            <div class="item-content">
              <el-select v-model="item.operator" size="small">
                <el-option label="等于" value="Equal"></el-option>
                <el-option label="存在" value="Exists"></el-option>
              </el-select>
            </div>
          </el-col>
          <el-col :span="4" style="padding-right: 10px;">
            <div class="item-content">
              <el-input v-model="item.value" size="small" placeholder=""></el-input>
            </div>
          </el-col>
          <el-col :span="4" style="padding-right: 10px;">
            <div class="item-content">
              <el-select v-model="item.effect" size="small" >
                <el-option label="全部" value=""></el-option>
                <el-option label="不调度" value="NoSchedule"></el-option>
                <el-option label="倾向于不调度" value="PreferNoSchedule"></el-option>
                <el-option label="不执行" value="NoExecute"></el-option>
              </el-select>
            </div>
          </el-col>
          <el-col :span="2" style="padding-right: 10px;">
            <div class="item-content">
              <el-input v-model="item.tolerationSeconds" size="small"
                placeholder="" controls-position="right">
                <template slot="suffix">
                  秒
                </template> 
              </el-input>
            </div>
          </el-col>
          <el-col :span="3">
            <el-button plain size="mini" style="padding-left: 10px; padding-right: 10px;" 
              @click="podSpec.tolerations.splice(idx, 1)" icon="el-icon-minus"></el-button>
          </el-col>
        </el-row>
        <el-button plain size="mini" @click="podSpec.tolerations.push({})" icon="el-icon-plus"></el-button>
      </el-form-item>
      <el-form-item label="节点亲和性" prop="nodeAffinity">
        <el-card class="box-card" style="margin-bottom: 20px;" v-for="(aff, idx) in podSpec.affinity.nodeAffinity" 
          :key="idx" :body-style="{padding: '20px'}">
          <el-row style="margin-bottom: 10px;">
            <el-col :span="aff.type === 'preferred' ? 9 : 22">
              <div class="item-header">
                优先级
              </div>
              <el-select v-model="aff.type" size="small" class="input-class">
                <el-option label="必须" value="required"></el-option>
                <el-option label="最好" value="preferred"></el-option>
              </el-select>
            </el-col>
            <el-col :span="13" v-if="aff.type === 'preferred'">
              <div class="item-header">
                <span  class="border-span-content">*</span>权重
              </div>
              <el-input-number controls-position="right" v-model="aff.weigth" size="small" placeholder="" :max="100" :min="1"
              ></el-input-number>
            </el-col>
            <el-col :span="2">
              <el-button style="float: right; padding: 3px 0" type="text"
                @click="podSpec.affinity.nodeAffinity.splice(idx, 1)">删除</el-button>
            </el-col>
          </el-row>
          <el-row style="margin-bottom: -15px;" v-if="aff.nodeSelectorTerms.length > 0">
            <el-col :span="3">
              <div class="border-span-header">
                属性
              </div>
            </el-col>
            <el-col :span="7">
              <div class="border-span-header">
                键
              </div>
            </el-col>
            <el-col :span="3">
              <div class="border-span-header">
                运算符
              </div>
            </el-col>
            <el-col :span="10">
              <div class="border-span-header">
                值
              </div>
            </el-col>
          </el-row>
          <el-row style="padding-top: 5;" v-for="(item, t_idx) in aff.nodeSelectorTerms" :key="t_idx">
            <el-col :span="3" style="padding-right: 10px;">
              <div class="item-content">
                <el-select v-model="item.type" size="small" >
                  <el-option label="标签" value="label"></el-option>
                  <el-option label="字段" value="field"></el-option>
                </el-select>
              </div>
            </el-col>
            <el-col :span="7" style="padding-right: 10px;">
              <div class="item-content">
                <el-input v-model="item.key" size="small" placeholder=""></el-input>
              </div>
            </el-col>
            <el-col :span="3" style="padding-right: 10px;">
              <div class="item-content">
                <el-select v-model="item.operator" size="small">
                  <el-option label="In" value="In"></el-option>
                  <el-option label="NotIn" value="NotIn"></el-option>
                  <el-option label="Exists" value="Exists"></el-option>
                  <el-option label="DoesNotExist" value="DoesNotExist"></el-option>
                  <el-option label="Gt(>)" value="Gt"></el-option>
                  <el-option label="Lt(<)" value="Lt"></el-option>
                </el-select>
              </div>
            </el-col>
            <el-col :span="7" style="padding-right: 10px;">
              <div class="item-content">
                <el-input v-model="item.values" size="small" placeholder="" :disabled="['Exists', 'DoesNotExist'].indexOf(item.operator) >= 0"></el-input>
              </div>
            </el-col>
            <el-col :span="3">
              <el-button plain size="mini" style="padding-left: 10px; padding-right: 10px;" 
                @click="aff.nodeSelectorTerms.splice(t_idx, 1)" icon="el-icon-minus"></el-button>
            </el-col>
          </el-row>
          <el-button style="padding: 3px 0" type="text"
            @click="aff.nodeSelectorTerms.push({type: 'label', operator: 'In'})">添加规则</el-button>
        </el-card>
        <el-button plain size="mini" @click="podSpec.affinity.nodeAffinity.push({type: 'required', weight: 1, nodeSelectorTerms: []})" icon="el-icon-plus"></el-button>
      </el-form-item>
      <el-form-item label="Pod亲和性" prop="podAffinity">
        <el-card class="box-card" style="margin-bottom: 20px;" v-for="(aff, idx) in podSpec.affinity.podAffinity" 
          :key="idx" :body-style="{padding: '20px'}">
          <el-row style="margin-bottom: 10px;">
            <el-col :span="9">
              <div class="item-header">
                优先级
              </div>
              <div style="padding-right: 18px;" class="pnClass">
                <el-select v-model="aff.type" size="small">
                  <el-option label="必须" value="required"></el-option>
                  <el-option label="最好" value="preferred"></el-option>
                </el-select>
              </div>
            </el-col>
            <el-col :span="13" v-if="aff.type === 'preferred'">
              <div class="item-header">
                <span  class="border-span-content">*</span>权重
              </div>
              <el-input-number controls-position="right" v-model="aff.weigth" size="small" placeholder="" max="100" min="1"
              ></el-input-number>
            </el-col>
            <el-col :span="aff.type === 'preferred' ? 2 : 15">
              <el-button style="float: right; padding: 3px 0" type="text"
                @click="podSpec.affinity.podAffinity.splice(idx, 1)">删除</el-button>
            </el-col>
          </el-row>
          <el-row style="margin-bottom: 10px;">
            <el-col :span="9">
              <div class="item-header">
                <span  class="border-span-content">*</span>拓扑键
              </div>
              <div style="padding-right: 18px;">
                <el-input v-model="aff.podAffinityTerm.topologyKey" size="small" placeholder=""></el-input>
              </div>
            </el-col>
            <el-col :span="9">
              <div class="item-header">
                命名空间
              </div>
              <div class="pnClass" style="padding-right: 18px;">
                <el-select v-model="aff.podAffinityTerm.namespaces" multiple placeholder="默认为空，表示当前命名空间" size="small">
                  <el-option
                    v-for="n in namespaces"
                    :key="n.name"
                    :label="n.name"
                    :value="n.name">
                  </el-option>
                </el-select>
              </div>
            </el-col>
          </el-row>
          <el-row style="margin-bottom: -15px;" v-if="aff.podAffinityTerm.labelSelector.length > 0">
            <el-col :span="7">
              <div class="border-span-header">
                标签键
              </div>
            </el-col>
            <el-col :span="3">
              <div class="border-span-header">
                运算符
              </div>
            </el-col>
            <el-col :span="14">
              <div class="border-span-header">
                标签值
              </div>
            </el-col>
          </el-row>
          <el-row style="padding-top: 5;" v-for="(item, t_idx) in aff.podAffinityTerm.labelSelector" :key="t_idx">
            <el-col :span="7" style="padding-right: 10px;">
              <div class="item-content">
                <el-input v-model="item.key" size="small" placeholder=""></el-input>
              </div>
            </el-col>
            <el-col :span="3" style="padding-right: 10px;">
              <div class="item-content">
                <el-select v-model="item.operator" size="small">
                  <el-option label="Equal" value="Equal"></el-option>
                  <el-option label="In" value="In"></el-option>
                  <el-option label="NotIn" value="NotIn"></el-option>
                  <el-option label="Exists" value="Exists"></el-option>
                  <el-option label="DoesNotExist" value="DoesNotExist"></el-option>
                </el-select>
              </div>
            </el-col>
            <el-col :span="7" style="padding-right: 10px;">
              <div class="item-content">
                <el-input v-model="item.values" size="small" placeholder=""></el-input>
              </div>
            </el-col>
            <el-col :span="3" style="padding-right: 10px;">
              <el-button plain size="mini" style="padding-left: 10px; padding-right: 10px;" 
                @click="aff.podAffinityTerm.labelSelector.splice(t_idx, 1)" icon="el-icon-minus"></el-button>
            </el-col>
          </el-row>
          <el-button style="padding: 3px 0" type="text"
            @click="aff.podAffinityTerm.labelSelector.push({operator: 'Equal'})">添加规则</el-button>
        </el-card>
        <el-button plain size="mini" @click="podSpec.affinity.podAffinity.push({type: 'required', weight: 1, podAffinityTerm: {labelSelector: [], namespaces: []}})" icon="el-icon-plus"></el-button>
      </el-form-item>
      <el-form-item label="Pod反亲和性" prop="podAntiAffinity">
        <el-card class="box-card" style="margin-bottom: 20px;" v-for="(aff, idx) in podSpec.affinity.podAntiAffinity" 
          :key="idx" :body-style="{padding: '20px'}">
          <el-row style="margin-bottom: 10px;">
            <el-col :span="9">
              <div class="item-header">
                优先级
              </div>
              <div style="padding-right: 18px;" class="pnClass">
                <el-select v-model="aff.type" size="small">
                  <el-option label="必须" value="required"></el-option>
                  <el-option label="最好" value="preferred"></el-option>
                </el-select>
              </div>
            </el-col>
            <el-col :span="13" v-if="aff.type === 'preferred'">
              <div class="item-header">
                <span  class="border-span-content">*</span>权重
              </div>
              <el-input-number controls-position="right" v-model="aff.weigth" size="small" placeholder="" max="100" min="1"
              ></el-input-number>
            </el-col>
            <el-col :span="aff.type === 'preferred' ? 2 : 15">
              <el-button style="float: right; padding: 3px 0" type="text"
                @click="podSpec.affinity.podAntiAffinity.splice(idx, 1)">删除</el-button>
            </el-col>
          </el-row>
          <el-row style="margin-bottom: 10px;">
            <el-col :span="9">
              <div class="item-header">
                <span  class="border-span-content">*</span>拓扑键
              </div>
              <div style="padding-right: 18px;">
                <el-input v-model="aff.podAffinityTerm.topologyKey" size="small" placeholder=""></el-input>
              </div>
            </el-col>
            <el-col :span="9">
              <div class="item-header">
                命名空间
              </div>
              <div class="pnClass" style="padding-right: 18px;">
                <el-select v-model="aff.podAffinityTerm.namespaces" multiple placeholder="默认为空，表示当前命名空间" size="small">
                  <el-option
                    v-for="n in namespaces"
                    :key="n.name"
                    :label="n.name"
                    :value="n.name">
                  </el-option>
                </el-select>
              </div>
            </el-col>
          </el-row>
          <el-row style="margin-bottom: -15px;" v-if="aff.podAffinityTerm.labelSelector.length > 0">
            <el-col :span="7">
              <div class="border-span-header">
                标签键
              </div>
            </el-col>
            <el-col :span="3">
              <div class="border-span-header">
                运算符
              </div>
            </el-col>
            <el-col :span="14">
              <div class="border-span-header">
                标签值
              </div>
            </el-col>
          </el-row>
          <el-row style="padding-top: 5;" v-for="(item, t_idx) in aff.podAffinityTerm.labelSelector" :key="t_idx">
            <el-col :span="7" style="padding-right: 10px;">
              <div class="item-content">
                <el-input v-model="item.key" size="small" placeholder=""></el-input>
              </div>
            </el-col>
            <el-col :span="3" style="padding-right: 10px;">
              <div class="item-content">
                <el-select v-model="item.operator" size="small">
                  <el-option label="Equal" value="Equal"></el-option>
                  <el-option label="In" value="In"></el-option>
                  <el-option label="NotIn" value="NotIn"></el-option>
                  <el-option label="Exists" value="Exists"></el-option>
                  <el-option label="DoesNotExist" value="DoesNotExist"></el-option>
                </el-select>
              </div>
            </el-col>
            <el-col :span="7" style="padding-right: 10px;">
              <div class="item-content">
                <el-input v-model="item.values" size="small" placeholder="" :disabled="['Exists', 'DoseNotExist'].indexOf(item.operator)"></el-input>
              </div>
            </el-col>
            <el-col :span="3" style="padding-right: 10px;">
              <el-button plain size="mini" style="padding-left: 10px; padding-right: 10px;" 
                @click="aff.podAffinityTerm.labelSelector.splice(t_idx, 1)" icon="el-icon-minus"></el-button>
            </el-col>
          </el-row>
          <el-button style="padding: 3px 0" type="text"
            @click="aff.podAffinityTerm.labelSelector.push({operator: 'Equal'})">添加规则</el-button>
        </el-card>
        <el-button plain size="mini" @click="podSpec.affinity.podAntiAffinity.push({type: 'required', weight: 1, podAffinityTerm: {labelSelector: [], namespaces: []}})" icon="el-icon-plus"></el-button>
      </el-form-item>
    </el-form>
  </div>
</template>

<script>

export default {
  name: 'PodAffinity',
  components: {
  },
  data() {
    return {
      podSpec: this.template.spec.template.spec,
      rules: {
        'name': [{ required: true, message: ' ', trigger: ['blur', 'change'] },],
      },
      nodes: []
    }
  },
  props: ['template'],
  computed: {
  },
  methods: {
  }
}
</script>

<style scoped lang="scss">
.input-class {
  width: 300px;
}

.border-span-header {
  padding-bottom: 8px;
}

.border-span-content {
  color: #F56C6C; 
  margin-right: 4px;
}
.box-card {
  font-size: 14px;
  color: #99a9bf
}
</style>

<style lang="scss">

</style>