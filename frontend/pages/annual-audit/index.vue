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

    <div class="flex flex-col md:flex-row gap-4 mb-4 bg-white dark:bg-gray-900 p-4 rounded-xl shadow-sm border border-gray-100 dark:border-gray-800">
      
      <UInput 
        v-model="searchCode" 
        icon="i-heroicons-magnifying-glass" 
        placeholder="Cari Code / Nama Aktivitas..." 
        class="flex-1"
        clearable
      />
      
      <USelectMenu 
        v-model="selectedDepartment" 
        :items="departmentOptions" 
        placeholder="Semua Departemen" 
        class="w-full md:w-60"
        clearable
      />
      
      <USelectMenu 
        v-model="selectedStatus" 
        :items="statusOptions" 
        placeholder="Semua Status" 
        class="w-full md:w-60"
        clearable
      />

      <UButton 
        v-if="searchCode || selectedDepartment || selectedStatus"
        icon="i-heroicons-x-mark" 
        color="neutral" 
        variant="ghost" 
        @click="clearFilters" 
        title="Reset Filters"
      />
    </div>

    <UCard class="rounded-xl shadow overflow-hidden" variant="soft" color="primary">
      <UTable
        :data="filteredPlans"    
        :columns="columns"
        :empty-state="{ icon: 'i-heroicons-circle-stack-20-solid', label: 'Belum ada data rencana audit.' }"
        class="w-full text-sm text-left"
      >
        <template #activity-cell="{ row }">
          <div class="py-2">
            <div class="font-bold text-gray-600">{{ row.original.code }}</div>
            <!-- <div class="font-medium text-gray-900 dark:text-white">{{row.original.name }}</div>
            <UBadge color="primary" variant="soft" size="md" class="mt-1">
              {{ row.original.type }}
            </UBadge> -->
          </div>
        </template>

        <template #department-cell="{ row }">
          <div class="flex gap-1 flex-wrap">
            <UBadge 
              v-for="dept in Array.from(new Set(row.original.activities.map(a => a.department)))" 
              :key="dept"
              color="neutral"
              variant="soft"
              size="md"
            >
              {{ dept }}
            </UBadge>
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

        <template #progress-cell="{ row }">
          <UProgress v-model="progressAudit" color="secondary" status />
        </template>

        <template #status-data="{ row }">
          <span 
            class="w-2.5 h-2.5 rounded-full inline-block"
            :class="getStatusColor(row.original.status)"
          />
        </template>

        <template #actions-cell="{ row }">
          <div class="flex">
            <UButton
              label="View"
              color="primary"
              variant="ghost"
              size="lg"
              @click="openViewModal(row.original)"
            />
            <h4> | </h4>
            <UButton
              label="Edit"
              color="primary"
              variant="ghost"
              size="lg"
              @click="handleEdit(row.original)"
            />
          </div>
        </template>

        <!-- <template #auditor-cell="{ row }">
          <div class="text-gray-700 dark:text-gray-300 flex items-center gap-1">
            <UIcon name="i-heroicons-users" class="w-4 h-4" />
            <span>{{ row.original.auditorCount }} Auditor</span>
          </div>
          <div class="text-xs text-gray-500 mt-0.5">@ {{ row.original.daysPerAuditor }} days</div>
        </template>

        <template #totalMandays-data="{ row }">
          <span class="font-bold text-gray-800 dark:text-gray-200">{{ row.original.totalMandays }}</span>
        </template>

        <template #supervisorName-data="{ row }">
          <span class="font-bold text-gray-800 dark:text-gray-200">{{ row.original.supervisorId || '-' }}</span>
        </template> -->

        <!-- <template #notes-cell="{ row }">
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
        </template> -->
      
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
                
                <div class="grid grid-cols-1 gap-6 md:grid-cols-2 mb-4">
                  <UFormField label="Status" size="lg">
                    <select v-model="form.status" class="input-field bg-white w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500">
                      <option value="Done">Done</option>
                      <option value="Work In Progress">Work In Progress</option>
                      <option value="Not Available">Not Available</option>
                    </select>
                  </UFormField>

                  <UFormField label="Activity Code" size="lg">  
                    <UInput 
                      v-model="form.code"
                      required 
                      type="text" 
                      placeholder="e.g. ASR-01"
                      class="w-full"
                    />
                  </UFormField>
                </div>

                <div class="space-y-4 mt-6">
                  <div 
                    v-for="(activity, index) in form.activities" 
                    :key="index" 
                    class="p-5 border border-gray-200 dark:border-gray-700 rounded-xl relative bg-gray-50 dark:bg-gray-800"
                  >
                    <div class="flex justify-between items-center mb-4 border-b border-gray-200 dark:border-gray-700 pb-2">
                      <h5 class="font-bold text-gray-700 dark:text-gray-300">Sub-Aktivitas {{ index + 1 }}</h5>
                      <UButton 
                        v-if="form.activities.length > 1"
                        icon="i-heroicons-trash" 
                        color="error" 
                        variant="ghost" 
                        size="sm"
                        @click="removeActivity(index)" 
                        title="Hapus aktivitas ini"
                      />
                    </div>

                    <div class="grid grid-cols-1 gap-6 md:grid-cols-3">
                      <UFormField label="Activity Name" size="lg">
                        <UInput 
                          v-model="activity.name" 
                          type="text" 
                          placeholder="e.g. Audit Operasional Div. Keuangan"
                          class="w-full"
                          required
                        />
                      </UFormField>

                      <UFormField label="Activity Type" size="lg">
                        <select v-model="activity.type" class="input-field bg-white w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500">
                          <option value="Assurance">Assurance</option>
                          <option value="Special Audit">Special Audit</option>
                          <option value="Specific Reason">Specific Reason</option>
                          <option value="Consulting Services">Consulting Services</option>
                          <option value="Investigation">Investigation</option>
                          <option value="Quality Assurance Review">Quality Assurance Review</option>
                          <option value="Follow-Up Audit">Follow-up Audit</option>
                        </select>
                      </UFormField>

                      <UFormField label="Department" size="lg">
                        <select v-model="activity.department" class="input-field bg-white w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500">
                          <option value="IT">IT</option>
                          <option value="Finance">Finance</option>
                          <option value="HR">HR</option>
                          <option value="Ops">Ops</option>
                        </select>
                      </UFormField>
                    </div>
                  </div>
                </div>

                <div class="flex justify-start pt-2">
                  <UButton 
                    label="Tambah Aktivitas" 
                    icon="i-heroicons-plus" 
                    color="primary" 
                    variant="soft" 
                    @click="addActivity" 
                  />
                </div>
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

    <Teleport to="body">
      <div v-if="showViewModal && selectedPlan" class="relative z-[9999]">
        <div class="fixed inset-0 bg-gray-900/80 transition-opacity" @click="closeViewModal"></div>
        <div class="fixed inset-0 z-10 w-screen overflow-y-auto">
          <div class="flex min-h-full items-center justify-center p-4">
            
            <div class="relative bg-white dark:bg-gray-900 rounded-xl shadow-2xl w-full max-w-5xl flex flex-col max-h-[95vh] overflow-y-auto p-8">
              
              <div class="flex justify-between items-center mb-6 pb-4 border-b border-gray-200 dark:border-gray-700">
                <div class="flex items-center gap-4">
                  <h2 class="text-2xl font-bold text-gray-800 dark:text-white">
                    [ {{ selectedPlan.code }} ] {{ selectedPlan.name }}
                  </h2>
                  <div class="flex items-center gap-2">
                    <span class="w-4 h-4 rounded-full" :class="getStatusColor(selectedPlan.status)"></span>
                    <span class="font-bold text-lg text-gray-800 dark:text-white">{{ selectedPlan.status }}</span>
                  </div>
                </div>
                
                <div class="flex items-center gap-6">
                  <button class="text-red-500 font-bold hover:underline text-sm" @click="handleDelete(selectedPlan.id)">Hapus</button>
                  <UButton 
                    label="Edit Data" 
                    color="warning" 
                    icon="i-heroicons-pencil-square" 
                    size="md"
                    class="font-bold"
                    @click="handleEditFromView(selectedPlan)" 
                  />
                  <button @click="closeViewModal" class="text-gray-400 hover:text-gray-600 text-3xl ml-2">&times;</button>
                </div>
              </div>

              <div class="space-y-6">
                
                <div class="border border-gray-200 dark:border-gray-700 rounded-lg p-6">
                  <h3 class="text-lg font-bold text-gray-800 dark:text-white mb-6">Status Pelaksanaan</h3>
                  <div class="flex items-center gap-8">
                    <span class="font-bold text-gray-700 dark:text-gray-300 w-32">Progress</span>
                    <div class="flex-1 max-w-xl">
                      <UProgress v-model="progressAudit" color="secondary" class="h-3" />
                    </div>
                    <span class="text-secondary-600 font-bold">50 %</span>
                  </div>
                </div>

                <div class="border border-gray-200 dark:border-gray-700 rounded-lg p-6">
                  <h3 class="text-lg font-bold text-gray-800 dark:text-white mb-6">1. Detil Aktivitas</h3>
                  
                  <div class="space-y-6">
                    <div class="flex items-center pb-4 border-b border-gray-100 dark:border-gray-800">
                      <span class="font-bold text-gray-700 dark:text-gray-300 w-48">Kode Aktivitas</span>
                      <span class="font-black text-primary-600 dark:text-primary-400">{{ selectedPlan.code }}</span>
                    </div>

                    <div 
                      v-for="(activity, index) in (selectedPlan.activities || [])" 
                      :key="index"
                      class="p-4 bg-gray-50 dark:bg-gray-800/50 rounded-lg space-y-3 border border-gray-100 dark:border-gray-800"
                    >
                      <div class="flex items-center justify-between border-b border-gray-200 dark:border-gray-700 pb-2 mb-2">
                        <span class="text-xs font-black uppercase text-gray-500 tracking-wider">
                          Sub-Aktivitas {{ Number(index) + 1 }}
                        </span>
                      </div>

                      <div class="flex items-start">
                        <span class="font-bold text-gray-600 dark:text-gray-400 w-44 text-sm">Nama Aktivitas</span>
                        <span class="font-semibold text-gray-800 dark:text-white flex-1">{{ activity.name }}</span>
                      </div>

                      <div class="flex items-center">
                        <span class="font-bold text-gray-600 dark:text-gray-400 w-44 text-sm">Kategori</span>
                        <UBadge size="xs" color="primary" variant="subtle" class="font-bold">
                          {{ activity.type }}
                        </UBadge>
                      </div>

                      <div class="flex items-center">
                        <span class="font-bold text-gray-600 dark:text-gray-400 w-44 text-sm">Departemen</span>
                        <span class="font-semibold text-gray-800 dark:text-white flex items-center gap-1">
                          🏢 {{ activity.department }}
                        </span>
                      </div>
                    </div>

                    <div v-if="!selectedPlan.activities?.length" class="text-center py-4 text-gray-400 italic">
                      Tidak ada detail aktivitas yang terdaftar.
                    </div>
                  </div>
                </div>

                <div class="border border-gray-200 dark:border-gray-700 rounded-lg p-6">
                  <h3 class="text-lg font-bold text-gray-800 dark:text-white mb-6">2. Timeline</h3>
                  <div class="space-y-4">
                    <div class="flex items-center"><span class="font-bold text-gray-700 dark:text-gray-300 w-48">Tahun</span><span class="font-semibold text-gray-800 dark:text-white">{{ selectedPlan.year }}</span></div>
                    <div class="flex items-center"><span class="font-bold text-gray-700 dark:text-gray-300 w-48">Distribusi Kuartal</span><span class="font-semibold text-gray-800 dark:text-white">{{ selectedPlan.quarters?.map((q: any) => `[ ${q} ]`).join(' ') || '-' }}</span></div>
                    <div class="flex items-center"><span class="font-bold text-gray-700 dark:text-gray-300 w-48">Pelaksanaan Bulan</span><span class="font-semibold text-gray-800 dark:text-white">{{ selectedPlan.selectedMonths?.slice().sort((a: number, b: number) => a - b).map((m: number) => `[ ${monthsList[m]} ]`).join(' ') || '-' }}</span></div>
                  </div>
                </div>

                <div class="border border-gray-200 dark:border-gray-700 rounded-lg p-6">
                  <h3 class="text-lg font-bold text-gray-800 dark:text-white mb-6">3. Auditor Resources</h3>
                  <div class="space-y-4">
                    <div class="flex items-center"><span class="font-bold text-gray-700 dark:text-gray-300 w-48">Supervisor</span><span class="font-semibold text-gray-800 dark:text-white">{{ getSupervisorName(selectedPlan.supervisorId) }}</span></div>
                    <div class="flex items-center"><span class="font-bold text-gray-700 dark:text-gray-300 w-48">Alokasi Waktu</span><span class="font-semibold text-gray-800 dark:text-white">👥 {{ selectedPlan.auditorCount }} Auditor ⏱️ {{ selectedPlan.daysPerAuditor }} Hari Durasi</span></div>
                    <div class="flex items-center"><span class="font-bold text-gray-700 dark:text-gray-300 w-48">Total Mandays</span><span class="font-semibold text-gray-800 dark:text-white">🔥 {{ selectedPlan.auditorCount * selectedPlan.daysPerAuditor }} Mandays</span></div>
                    <div class="flex items-center"><span class="font-bold text-gray-700 dark:text-gray-300 w-48">Estimasi Kapasitas</span><span class="font-semibold text-gray-800 dark:text-white">[🟢 Optimal (60-80%)] Total Load: 0.6% dari Kap Tahunan</span></div>
                  </div>
                </div>

                <div class="border border-gray-200 dark:border-gray-700 rounded-lg p-6">
                  <h3 class="text-lg font-bold text-gray-800 dark:text-white mb-6">4. Additional Notes</h3>
                  <p class="font-semibold text-gray-800 dark:text-white">{{ selectedPlan.notes || '-' }}</p>
                </div>

              </div>

            </div>
          </div>
        </div>
      </div>
    </Teleport>

  </div>
</template>

<script setup lang="ts">
import type { TableColumn } from '@nuxt/ui'
import { useAnnualPlanStore } from '~/stores/annual-audit'
import { AnnualAuditPlanStatus, AuditDepartment, AuditType, type AnnualAuditPlan, type AnnualPlanForm } from '~/types/audit'

const store = useAnnualPlanStore()
const showModal = ref(false)
const progressAudit = ref(50);

// Constants
const monthsList = ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec']

const columns: TableColumn<AnnualAuditPlan>[] = [
  { accessorKey: 'activity', header: 'Activity' },
  { accessorKey: 'department', header: 'Department' },
  { accessorKey: 'timeline', header: 'Timeline' },
  { accessorKey: 'progress', header: 'Progress' },
  { accessorKey: 'status', header: 'Status' },
  { accessorKey: 'actions', header: 'Actions' },
  // { accessorKey: 'auditor', header: 'Total Auditor' },
  // { accessorKey: 'totalMandays', header: 'Mandays Duration' },
  // { accessorKey: 'supervisorName', header: 'Supervisor' },
  // { accessorKey: 'notes', header: 'Notes' },
]

const isEditing = ref(false)
const editingId = ref<string | null>(null)
  // --- STATE UNTUK MODAL VIEW ---
const showViewModal = ref(false)
const selectedPlan = ref<any>(null)

// --- STATE UNTUK FILTER ---
const searchCode = ref('')
const selectedDepartment = ref<string | undefined>(undefined)
const selectedStatus = ref<string | undefined>(undefined)

// --- OPSI UNTUK DROPDOWN FILTER ---
const statusOptions = ['Done', 'Work In Progress', 'Not Available']
// Asumsi Anda menggunakan Object.values(AuditDepartment) jika berupa enum, atau array manual
const departmentOptions = ['Finance', 'IT', 'Operations', 'HR', 'Compliance'] // Sesuaikan dengan enum Anda

// --- COMPUTED: FILTER DATA ---
// --- 3. UPDATE FILTERED PLANS ---
const filteredPlans = computed(() => {
  return store.plans.filter(plan => {
    
    // 1. Filter by Code OR ANY Activity Name
    const matchCode = !searchCode.value || 
      plan.code.toLowerCase().includes(searchCode.value.toLowerCase()) || 
      plan.activities.some(act => act.name.toLowerCase().includes(searchCode.value.toLowerCase()))
      
    // 2. Filter by Department (Cek apakah ada sub-aktivitas dengan departemen tsb)
    const matchDept = !selectedDepartment.value || 
      plan.activities.some(act => act.department === selectedDepartment.value)
    
    // 3. Filter by Status (Tetap sama)
    const matchStatus = !selectedStatus.value || plan.status === selectedStatus.value

    return matchCode && matchDept && matchStatus
  })
})

// Fungsi Reset Filter
const clearFilters = () => {
  searchCode.value = ''
  selectedDepartment.value = undefined
  selectedStatus.value = undefined
}

// --- ACTIONS UNTUK MODAL VIEW ---
const openViewModal = (plan: any) => {
  selectedPlan.value = plan
  showViewModal.value = true
}

const closeViewModal = () => {
  showViewModal.value = false
  setTimeout(() => { selectedPlan.value = null }, 200) // Delay agar transisi tidak flicker
}

const handleEditFromView = (plan: any) => {
  closeViewModal() // Tutup modal detail
  handleEdit(plan) // Buka modal edit bawaan form
}

const getSupervisorName = (id: string) => {
  if (!id) return '-'
  const supervisor = store.supervisors.find(s => s.id === id)
  return supervisor ? supervisor.name : id
}

const getStatusColor = (status: string) => {
  switch (status) {
    case 'Done': return 'bg-green-500'
    case 'Work In Progress': return 'bg-amber-500'
    case 'Not Available': return 'bg-gray-400'
    default: return 'bg-gray-200'
  }
}

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
  activities: [
    { name: '', type: AuditType.ASSURANCE, department: AuditDepartment.IT }
  ],
  // name: '',
  // type: AuditType.ASSURANCE,
  // department: AuditDepartment.IT,
  status: AnnualAuditPlanStatus.NOT_AVAILABLE,
  selectedMonths: [],
  auditorCount: 2,
  daysPerAuditor: 5,
  supervisorId: '',
  notes: '',
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

const addActivity = () => {
  form.activities.push({
    name: '',
    type: AuditType.ASSURANCE,
    department: AuditDepartment.IT,
  })
}

const removeActivity = (index: number) => {
  // Cegah penghapusan jika hanya tersisa 1 aktivitas
  if (form.activities.length > 1) {
    form.activities.splice(index, 1)
  }
}

const openModal = () => {
  isEditing.value = false
  editingId.value = null
  
  // Reset Form
  Object.assign(form, {
    code: '',
    activities: [ // Reset array activities kembali ke 1 baris kosong
      { name: '', type: AuditType.ASSURANCE, department: AuditDepartment.IT }
    ],
    // name: '',
    // type: AuditType.ASSURANCE,
    // department: AuditDepartment.IT,
    status: AnnualAuditPlanStatus.NOT_AVAILABLE,
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
    activities: plan.activities.map((act: any) => ({ ...act })),
    // name: plan.name,
    // type: plan.type,
    // department: plan.department,
    status: plan.status,
    selectedMonths: [...plan.selectedMonths], // Gunakan spread agar tidak reaktif terhubung langsung
    auditorCount: plan.auditorCount,
    daysPerAuditor: plan.daysPerAuditor,
    supervisorId: plan.supervisorId,
    notes: plan.notes || '',
    isActive: plan.isActive,
    year: plan.year
  })
  
  showModal.value = true
}

const handleDelete = (id: string | undefined) => {
  if (!id) return
    try {
      // 2. Panggil fungsi hapus di store
      store.deletePlan(id)
      
      // 3. Tutup modal detail setelah berhasil menghapus
      closeViewModal()
      
      // 4. (Opsional) Tambahkan notifikasi sukses menggunakan Nuxt UI Toast
      // toast.add({ title: 'Sukses', description: 'Rencana audit telah dihapus.', color: 'green' })
      
    } catch (error) {
      alert('Gagal menghapus data: ' + error)
    }
}

</script>