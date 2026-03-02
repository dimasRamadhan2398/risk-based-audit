<template>
  <div class="p-6 max-w-7xl mx-auto space-y-6">
    <div class="flex justify-between items-center">
      <h1 class="text-2xl font-bold text-gray-900 dark:text-white"> 
        Annual Audit Plan
      </h1>
      <UButton
        label="New Audit Plan" 
        @click="openModal()"
        color="primary" 
        class="px-4 py- font-bold shadow-lg flex gap-2"
        icon="add"
        >
      </UButton>
    </div>

    <UCard class="rounded-xl shadow overflow-hidden" variant="soft" color="primary">
      <UTable
        :data="store.plans"    
        :columns="columns"
        :empty-state="{ icon: 'i-heroicons-circle-stack-20-solid', label: 'Belum ada data rencana audit.' }"
        class="w-full text-sm text-left"
      >
        <template #activity-cell="{ row }">
          <div class="py-2">
            <div class="font-bold text-gray-600">{{ row.original.code }}</div>
            <div class="font-medium text-gray-900 dark:text-white">{{row.original.name }}</div>
            <UBadge color="primary" variant="soft" size="md" class="mt-1">
              {{ row.original.type }}
            </UBadge>
          </div>
          <div class="max-w-[250px] text-sm text-gray-600 dark:text-gray-400">
            <template v-if="!isDetailLongText(row.original.detail!)">
              <span class="italic">{{ row.original.detail || '-' }}</span>
            </template>

            <template v-else>
              <span class="italic whitespace-normal break-words">
                {{ expandedDetailRows.has(row.original.id!) 
                  ? row.original.detail 
                  : row.original.detail?.slice(0, 50) + '...' 
                }}
              </span>
              
              <UButton
                :label="expandedDetailRows.has(row.original.id!) ? 'Show Less' : 'Read More'"
                variant="link"
                size="xs"
                :padded="false"
                color="primary"
                class="ml-1 font-bold underline"
                @click="toggleDetailReadMore(row.original.id!)"
              />
            </template>
          </div>
        </template>

        <template #timeline-cell="{ row }">
          <div class="font-bold text-primary-600 mr-1">{{ row.original.year }}</div>
          <UBadge v-for="q in row.original.quarters" :key="q" color="primary" variant="subtle" size="md">
            {{ q }}
          </UBadge>
          <div class="flex gap-1 flex-wrap mt-1">
            <UBadge 
              v-for="idx in row.original.selectedMonths.slice().sort((a, b) => a - b)" 
              :key="idx"
              color="primary" 
              variant="outline" 
              size="md"
            >
              {{ monthsList[idx] }}
            </UBadge>
          </div>
        </template>

        <template #auditor-cell="{ row }">
          <div class="text-gray-700 dark:text-gray-300 flex items-center gap-1">
            <UIcon name="i-heroicons-users" class="w-4 h-4" />
            <span>{{ row.original.auditorCount }} Auditor</span>
          </div>
          <!-- <div class="text-xs text-gray-500 mt-0.5">@ {{ row.original.daysPerAuditor }} days</div> -->
        </template>

        <template #totalMandays-data="{ row }">
          <span class="font-bold text-gray-800 dark:text-gray-200">{{ row.original.totalMandays }}</span>
        </template>

        <template #supervisorName-data="{ row }">
          <span class="font-bold text-gray-800 dark:text-gray-200">{{ row.original.supervisorId || '-' }}</span>
        </template>

        <template #notes-cell="{ row }">
          <div class="max-w-[250px] text-sm text-gray-600 dark:text-gray-400">
            <template v-if="!isNotesLongText(row.original.notes!)">
              <span class="italic">{{ row.original.notes || '-' }}</span>
            </template>

            <template v-else>
              <span class="italic whitespace-normal break-words">
                {{ expandedNotesRows.has(row.original.id!) 
                  ? row.original.notes 
                  : row.original.notes?.slice(0, 50) + '...' 
                }}
              </span>
              
              <UButton
                :label="expandedNotesRows.has(row.original.id!) ? 'Show Less' : 'Read More'"
                variant="link"
                size="xs"
                :padded="false"
                color="primary"
                class="ml-1 font-bold underline"
                @click="toggleNotesReadMore(row.original.id!)"
              />
            </template>
          </div>
        </template>

        <template #actions-cell="{ row }">
          <div>
            <UButton
              label="Edit"
              icon="i-heroicons-pencil-square"
              color="primary"
              variant="ghost"
              size="lg"
              @click="handleEdit(row.original)"
            />
          </div>
        </template>
      
      </UTable>
    </UCard>

    <Teleport to="body">
      <div v-if="showModal" class="relative z-[9999]">
        <div class="fixed inset-0 bg-gray-900/80 transition-opacity"></div>
        <div class="fixed inset-0 z-10 w-screen overflow-y-auto">
          <div class="flex min-h-full items-center justify-center p-4">
          <UForm @submit.prevent="handleSubmit">
            <div class="relative bg-secondary-50 dark:bg-secondary-800 rounded-xl shadow-2xl w-full max-w-4xl border border-secondary-200 dark:border-secondary-700 flex flex-col max-h-[90vh]">
              
              <div class="px-6 py-4 border-b border-secondary-200 dark:border-secondary-700 bg-secondary-50 dark:bg-secondary-900 rounded-t-xl flex justify-between items-center">
                <UIcon name="charter" class=" text-primary-500" size="32"></UIcon>
                <h3 class="text-lg font-bold text-secondary-900 dark:text-white">Annual Audit Form</h3>
                <UIcon name="close" @click="closeModal" class="text-primary-400 hover:text-primary-600 text-2xl">&times;</UIcon>
              </div>

              
              <div class="p-6 overflow-y-auto space-y-8">  
                <div class="space-y-4">
                  <h4 class="text-sm uppercase tracking-wide text-primary-500 font-bold border-b pb-2">1. Activity Detail</h4>
                  <div class="grid grid-cols-1 md:grid-cols-3">
                    <UFormField 
                      label="Activity Code"
                      size="lg"
                    >  
                      <UInput 
                        v-model="form.code"
                        required 
                        type="text" 
                        placeholder="e.g. ASR-01" 
                      />
                    </UFormField>
                    <UFormField
                      label="Activity Name" 
                      size="lg"
                    >
                      <UInput 
                        v-model="form.name" 
                        type="text" 
                        placeholder="e.g. Audit Operasional Divisi Keuangan" 
                        required
                      />
                    </UFormField>
                    <UFormField
                      label="Type Name" 
                      size="lg"
                    >
                      <select v-model="form.type" class="input-field bg-white">
                        <option value="Assurance">Assurance</option>
                        <option value="Special Audit">Special Audit</option>
                        <option value="Specific Reason">Specific Reason</option>
                        <option value="Consulting Services">Consulting Services</option>
                        <option value="Investigation">Investigation</option>
                        <option value="Quality Assurance Review">Quality Assurance Review</option>
                        <option value="Follow-Up Audit">Follow-up Audit</option>
                      </select>
                    </UFormField>
                  </div>
                  <UFormField
                      label="Activity Detail"
                      size="lg"
                    >
                      <UTextarea 
                        v-model="form.detail" 
                        maxlength="500"
                        :rows="5"
                        class="w-full"
                        autoresize
                      ></UTextarea>
                      <div class="flex justify-between mt-1">
                        <span class="text-xs text-gray-400">Describe Audit Activity Plan here.</span>
                        <span class="text-xs text-gray-400">{{ form.detail!.length }}/500</span>
                      </div>
                    </UFormField>     
                </div>

                <div class="space-y-4">
                  <h4 class="text-sm uppercase tracking-wide text-primary-500 font-bold border-b pb-2">2. Timeline</h4>
                  
                  <UFormField
                      label="Year" 
                      size="lg"
                  >
                    <select v-model="form.year" class="input-field bg-white">
                      <option value="2026">2026</option>
                      <option value="2027">2027</option>
                      <option value="2028">2028</option>
                      <option value="2029">2029</option>
                      <option value="2030">2030</option>
                      <option value="2031">2031</option>
                      <option value="2032">2032</option>
                      <option value="2033">2033</option>
                      <option value="2034">2034</option>
                      <option value="2035">2035</option>
                      <option value="2036">2036</option>
                    </select>
                  </UFormField>
                  <UFormField
                    label="Select Months"
                    size="lg"
                  >
                    <div class="grid grid-cols-3 md:grid-cols-6 lg:grid-cols-12 gap-2 p-4">
                      <div v-for="(month, idx) in monthsList" :key="idx" 
                        @click="toggleMonth(idx)"
                        class="cursor-pointer border rounded-lg p-2 text-center text-xs font-semibold transition select-none"
                        :class="form.selectedMonths.includes(idx) ? 'bg-secondary-600 text-white border-secondary-600 shadow-md transform scale-105' : 'bg-gray-50 text-gray-500 hover:bg-gray-100'"
                      >
                        {{ month }}
                      </div>
                    </div>

                    <div v-if="scheduleWarning" class="flex items-center gap-2 text-warning-600 bg-warning-50 p-4 m-4 rounded-lg text-sm border border-warning-200">
                      <UIcon name="warning" class=" text-warning-500"></UIcon>
                      {{ scheduleWarning }}
                    </div>
                    <div v-if="quarterAlert" class="flex items-center gap-2 text-error-600 bg-red-50 p-4 m-4 rounded-lg text-sm border border-error-200">
                      <UIcon name="alert" class=" text-error-500"></UIcon>
                      {{ quarterAlert }}
                    </div>
                    
                    <div class="text-sm text-gray-600 bg-gray-50 p-4 m-4 rounded border">
                      <span class="font-bold">Distribusi Triwulan:</span> 
                      {{ computedQuarters.length ? computedQuarters.join(', ') : '-' }}
                    </div>
                  </UFormField>
                </div>

                <div class="space-y-4">
                  <h4 class="text-sm uppercase tracking-wide text-primary-500 font-bold border-b pb-2">3. Auditor</h4>
                  
                  <div class="grid grid-cols-1 md:grid-cols-3 gap-6 p-4 bg-primary-50 dark:bg-primary-700/30 rounded-lg border border-primary-200 dark:border-primary-700">
                    <UFormField
                      label="Number of Auditors (1-10)"
                      size="lg"
                    >
                      <UInput v-model.number="form.auditorCount" type="number" min="1" max="10" class="w-full"/>
                      <p class="text-xs text-gray-500 mt-1">Advice: High Risk min. 3 auditor</p>
                    </UFormField>
                    
                    <UFormField
                      label="Duration (Days)"
                      size="lg"
                    >
                      <UInput v-model.number="form.daysPerAuditor" type="number" min="1" />
                    </UFormField>

                    <UBadge class="bg-white dark:bg-gray-800 p-3 rounded border text-center flex flex-col justify-center">
                      <span class="text-xs text-gray-500 uppercase">Total Mandays</span>
                      <span class="text-2xl font-bold text-primary-600">{{ totalMandays }}</span>
                      <span class="text-xs text-gray-400">= {{ form.auditorCount }} person × {{ form.daysPerAuditor }} day</span>
                    </UBadge>
                  </div>

                  <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
                    <UFormField
                      label="Supervisor"
                      size="lg"
                    >
                      <select v-model="form.supervisorId" class="input-field bg-white">
                        <option value="" disabled>-- Choose Supervisor --</option>
                        <option v-for="s in store.supervisors" :key="s.id" :value="s.id">
                          {{ s.name }} (Workload: {{ s.workload }})
                        </option>
                      </select>
                      <p v-if="selectedSupervisor?.workload! > 6" class="text-xs text-error-500 mt-1 font-bold">
                        ⚠️ Warning: This Supervisor is supervised > 6 Activity!
                      </p>
                    </UFormField>

                    <UFormField>
                      <div class="flex justify-between items-center mb-1">
                        <label class="label mb-0">Team Estimation Capacity</label>
                        <span class="text-xs font-bold" :class="utilizationData.color === 'red' ? 'text-error-600' : 'text-success-600'">
                          {{ utilizationData.msg }}
                        </span>
                      </div>
                      <div class="w-full bg-gray-200 rounded-full h-2.5 dark:bg-gray-700">
                        <div class="h-2.5 rounded-full transition-all duration-500" 
                          :class="{
                            'bg-success-500': utilizationData.color === 'green',
                            'bg-warning-400': utilizationData.color === 'yellow',
                            'bg-error-600': utilizationData.color === 'red'
                          }"
                          :style="{ width: `${Math.min(utilizationData.percent, 100)}%` }"
                        ></div>
                      </div>
                      <p class="text-xs text-gray-400 mt-1">
                        Total Load: {{ utilizationData.percent.toFixed(1) }}% from Annual Capacity.
                      </p>
                    </UFormField>
                  </div>
                </div>

                <div class="space-y-4">
                  <h4 class="text-sm uppercase tracking-wide text-gray-500 font-bold border-b pb-2">4. Notes</h4>
                  <UFormField
                    label="Additional Note (Optional)"
                    size="lg"
                  >
                    <UTextarea 
                      v-model="form.notes" 
                      maxlength="500"
                      :rows="5"
                      placeholder="Example: High Priority - external recommendation..."
                      class="w-full"
                      autoresize
                    ></UTextarea>
                    <div class="flex justify-between mt-1">
                      <span class="text-xs text-gray-400">Use it for Special Audit / Investigation.</span>
                      <span class="text-xs text-gray-400">{{ form.notes!.length }}/500</span>
                    </div>
                  </UFormField>
                </div>
              </div>
            
              <div class="px-6 py-4 bg-secondary-50 dark:bg-secondary-900 border-t border-secondary-200 dark:border-secondary-700 rounded-b-xl flex justify-end gap-3">
                <UButton 
                  :label="isEditing ? 'Update Plan' : 'Save Plan'" 
                  color="primary" 
                  :disabled="!!quarterAlert || utilizationData.color === 'red'"
                  @click="handleSubmit" 
                />
              </div>
            
            </div>
          </UForm>  
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<script setup lang="ts">
import type { TableColumn } from '@nuxt/ui'
import { useAnnualPlanStore } from '~/stores/annual-audit'
import { AuditStatus, AuditType, type AnnualAuditPlan, type AnnualPlanForm } from '~/types/audit'

const store = useAnnualPlanStore()
const showModal = ref(false)

// Constants
const monthsList = ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec']

const columns: TableColumn<AnnualAuditPlan>[] = [
  // { accessorKey: 'code', header: 'Code' },
  // { accessorKey: 'type', header: 'Type' },
  // { accessorKey: 'name', header: 'Name' },
  // { accessorKey: 'tahun', header: 'Tahun' },
  // { accessorKey: 'bulan', header: 'Bulan' },
  { accessorKey: 'activity', header: 'Activity' },
  { accessorKey: 'timeline', header: 'Timeline' },
  { accessorKey: 'auditor', header: 'Total Auditor' },
  { accessorKey: 'totalMandays', header: 'Mandays Duration' },
  { accessorKey: 'supervisorName', header: 'Supervisor' },
  { accessorKey: 'notes', header: 'Notes' },
  { accessorKey: 'actions', header: 'Actions' }
]

const isEditing = ref(false)
const editingId = ref<string | null>(null)

// State untuk menyimpan ID baris yang sedang "Read More"
const expandedNotesRows = ref(new Set<string>())
  const expandedDetailRows = ref(new Set<string>())

const toggleNotesReadMore = (id: string) => {
  if (expandedNotesRows.value.has(id)) {
    expandedNotesRows.value.delete(id)
  } else {
    expandedNotesRows.value.add(id)
  }
}

const toggleDetailReadMore = (id: string) => {
  if (expandedDetailRows.value.has(id)) {
    expandedDetailRows.value.delete(id)
  } else {
    expandedDetailRows.value.add(id)
  }
}

// Helper untuk mengecek apakah teks lebih dari 100 karakter
const isNotesLongText = (text: string) => text && text.length > 100
const isDetailLongText = (text: string) => text && text.length > 100

// Form State
const form = reactive<AnnualPlanForm>({
  code: '',
  name: '',
  type: AuditType.ASSURANCE,
  detail: '',
  selectedMonths: [],
  auditorCount: 2,
  daysPerAuditor: 5,
  supervisorId: '',
  notes: '',
  status: AuditStatus.PLANNED,
  isActive: true,
  year: ''
})

// --- COMPUTED LOGIC (Real-time Validation) ---

// F-03: Auto Calculate Mandays
const totalMandays = computed(() => form.auditorCount * form.daysPerAuditor)

// F-03: Supervisor Check
const selectedSupervisor = computed(() => store.supervisors.find(s => s.id === form.supervisorId))

// F-03: Utilization Check
const utilizationData = computed(() => store.checkUtilization(totalMandays.value))

// F-02: Schedule Logic
const computedQuarters = computed(() => store.calculateQuarters(form.selectedMonths))

const scheduleWarning = computed(() => store.checkScheduleGaps(form.selectedMonths))

const quarterAlert = computed(() => {
  // F-02: Alert jika Q1 > 40% (Simplifikasi: jika user pilih semua bulan Q1 dan hanya sedikit bulan lain)
  const q1Count = form.selectedMonths.filter(m => m <= 2).length
  const totalCount = form.selectedMonths.length
  if (totalCount > 0 && (q1Count / totalCount) > 0.4 && totalCount > 3) {
    return "Beban kerja Triwulan I terlalu tinggi (>40%). Mohon ratakan jadwal."
  }
  return null
})

// --- ACTIONS ---

const toggleMonth = (idx: number) => {
  if (form.selectedMonths.includes(idx)) {
    form.selectedMonths = form.selectedMonths.filter(m => m !== idx)
  } else {
    form.selectedMonths.push(idx)
  }
}

const openModal = () => {
  isEditing.value = false
  editingId.value = null
  
  // Reset Form
  Object.assign(form, {
    code: '',
    name: '',
    type: AuditType.ASSURANCE,
    selectedMonths: [],
    auditorCount: 2,
    daysPerAuditor: 5,
    supervisorId: '',
    notes: '',
    year: ''
  })
  showModal.value = true
}

const closeModal = () => showModal.value = false

const handleSubmit = () => {
  // F-04: Final Validation
  if (form.selectedMonths.length === 0) {
    alert("⚠️ Wajib memilih minimal 1 bulan pelaksanaan.")
    return
  }
  if (!form.supervisorId) {
    alert("⚠️ Wajib memilih Supervisor.")
    return
  }

  if (isEditing.value && editingId.value) {
    // Mode EDIT
    store.updatePlan(editingId.value, { ...form })
    alert("Data Rencana Audit Berhasil Diperbarui!")
  } else {
    // Mode ADD
    store.addPlan({ ...form })
    alert("Data Rencana Audit Berhasil Disimpan!")
  }

  closeModal()
}

const handleEdit = (plan: any) => {
  isEditing.value = true
  editingId.value = plan.id
  
  // Isi form dengan data yang dipilih
  Object.assign(form, {
    code: plan.code,
    name: plan.name,
    type: plan.type,
    selectedMonths: [...plan.selectedMonths], // Gunakan spread agar tidak reaktif terhubung langsung
    auditorCount: plan.auditorCount,
    daysPerAuditor: plan.daysPerAuditor,
    supervisorId: plan.supervisorId,
    notes: plan.notes || '',
    status: plan.status,
    isActive: plan.isActive,
    year: plan.year
  })
  
  showModal.value = true
}

</script>