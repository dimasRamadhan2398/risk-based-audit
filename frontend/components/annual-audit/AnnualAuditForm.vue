<template>
    <Teleport to="body">
        <div v-if="store.showModal" class="relative z-[9999]">
        <div class="fixed inset-0 bg-gray-900/80 transition-opacity"></div>
        <div class="fixed inset-0 z-10 w-screen overflow-y-auto">
            <div class="flex min-h-full items-center justify-center p-4">
            <UForm @submit.prevent="store.handleSubmit">
            <div class="relative bg-secondary-50 dark:bg-secondary-800 rounded-xl shadow-2xl w-full max-w-4xl border border-secondary-200 dark:border-secondary-700 flex flex-col max-h-[90vh]">
                
                <div class="px-6 py-4 border-b border-secondary-200 dark:border-secondary-700 bg-secondary-50 dark:bg-secondary-900 rounded-t-xl flex justify-between items-center">
                <UIcon name="charter" class=" text-primary-500" size="32"></UIcon>
                <h3 class="text-lg font-bold text-secondary-900 dark:text-white">Annual Audit Form</h3>
                <UIcon name="close" @click="store.closeModal" class="text-primary-400 hover:text-primary-600 text-2xl">&times;</UIcon>
                </div>

                
                <div class="p-6 overflow-y-auto space-y-4">  
                <div class="space-y-4">
                <h4 class="text-sm uppercase tracking-wide text-primary-500 font-bold border-b pb-2">1. Activity Detail</h4>
                
                <div class="grid grid-cols-1 gap-6 md:grid-cols-2 mb-4">
                    <UFormField label="Status" size="lg">
                    <select v-model="store.form.status" class="input-field bg-white w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500">
                        <option value="Done">Done</option>
                        <option value="Work In Progress">Work In Progress</option>
                        <option value="Not Available">Not Available</option>
                    </select>
                    </UFormField>

                    <UFormField label="Activity Code" size="lg">  
                    <UInput 
                        v-model="store.form.code"
                        required 
                        type="text" 
                        placeholder="e.g. ASR-01"
                        class="w-full"
                    />
                    </UFormField>
                </div>

                <div class="space-y-4 mt-6">
                    <div 
                    v-for="(activity, index) in store.form.activities" 
                    :key="index" 
                    class="p-5 border border-gray-200 dark:border-gray-700 rounded-xl relative bg-gray-50 dark:bg-gray-800"
                    >
                    <div class="flex justify-between items-center mb-4 border-b border-gray-200 dark:border-gray-700 pb-2">
                        <h5 class="font-bold text-gray-700 dark:text-gray-300">Sub-Aktivitas {{ index + 1 }}</h5>
                        <UButton 
                        v-if="store.form.activities.length > 1"
                        icon="i-heroicons-trash" 
                        color="error" 
                        variant="ghost" 
                        size="sm"
                        @click="store.removeActivity(index)" 
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

                        <UFormField label="Activity Category" size="lg">
                        <select v-model="activity.category" class="input-field bg-white w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500">
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
                    @click="store.addActivity" 
                    />
                </div>
                </div>

                <div class="space-y-4">
                    <h4 class="text-sm uppercase tracking-wide text-primary-500 font-bold border-b pb-2">2. Timeline</h4>
                    
                    <UFormField
                        label="Year" 
                        size="lg"
                    >
                    <select v-model="store.form.year" class="input-field bg-white">
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
                        <div v-for="(month, idx) in store.monthsList" :key="idx" 
                        @click="store.toggleMonth(idx)"
                        class="cursor-pointer border rounded-lg p-2 text-center text-xs font-semibold transition select-none"
                        :class="store.form.selectedMonths.includes(idx) ? 'bg-secondary-600 text-white border-secondary-600 shadow-md transform scale-105' : 'bg-gray-50 text-gray-500 hover:bg-gray-100'"
                        >
                        {{ month }}
                        </div>
                    </div>

                    <div v-if="store.scheduleWarning" class="flex items-center gap-2 text-warning-600 bg-warning-50 p-4 m-4 rounded-lg text-sm border border-warning-200">
                        <UIcon name="warning" class=" text-warning-500"></UIcon>
                        {{ store.scheduleWarning }}
                    </div>
                    <div v-if="store.quarterAlert" class="flex items-center gap-2 text-error-600 bg-red-50 p-4 m-4 rounded-lg text-sm border border-error-200">
                        <UIcon name="alert" class=" text-error-500"></UIcon>
                        {{ store.quarterAlert }}
                    </div>
                    
                    <div class="text-sm text-gray-600 bg-gray-50 p-4 m-4 rounded border">
                        <span class="font-bold">Distribusi Triwulan:</span> 
                        {{ store.computedQuarters.length ? store.computedQuarters.join(', ') : '-' }}
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
                        <UInput v-model.number="store.form.auditorCount" type="number" min="1" max="10" class="w-full"/>
                        <p class="text-xs text-gray-500 mt-1">Advice: High Risk min. 3 auditor</p>
                    </UFormField>
                    
                    <UFormField
                        label="Duration (Days)"
                        size="lg"
                    >
                        <UInput v-model.number="store.form.daysPerAuditor" type="number" min="1" />
                    </UFormField>

                    <UBadge class="bg-white dark:bg-gray-800 p-3 rounded border text-center flex flex-col justify-center">
                        <span class="text-xs text-gray-500 uppercase">Total Mandays</span>
                        <span class="text-2xl font-bold text-primary-600">{{ store.totalMandays }}</span>
                        <span class="text-xs text-gray-400">= {{ store.form.auditorCount }} person × {{ store.form.daysPerAuditor }} day</span>
                    </UBadge>
                    </div>

                    <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
                    <UFormField
                        label="Supervisor"
                        size="lg"
                    >
                        <select v-model="store.form.supervisorId" class="input-field bg-white">
                        <option value="" disabled>-- Choose Supervisor --</option>
                        <option v-for="s in store.supervisors" :key="s.id" :value="s.id">
                            {{ s.name }} (Workload: {{ s.workload }})
                        </option>
                        </select>
                        <p v-if="store.selectedSupervisor?.workload! > 6" class="text-xs text-error-500 mt-1 font-bold">
                        ⚠️ Warning: This Supervisor is supervised > 6 Activity!
                        </p>
                    </UFormField>

                    <UFormField>
                        <div class="flex justify-between items-center mb-1">
                        <label class="label mb-0">Team Estimation Capacity</label>
                        <span class="text-xs font-bold" :class="store.utilizationData.color === 'red' ? 'text-error-600' : 'text-success-600'">
                            {{ store.utilizationData.msg }}
                        </span>
                        </div>
                        <div class="w-full bg-gray-200 rounded-full h-2.5 dark:bg-gray-700">
                        <div class="h-2.5 rounded-full transition-all duration-500" 
                            :class="{
                            'bg-success-500': store.utilizationData.color === 'green',
                            'bg-warning-400': store.utilizationData.color === 'yellow',
                            'bg-error-600': store.utilizationData.color === 'red'
                            }"
                            :style="{ width: `${Math.min(store.utilizationData.percent, 100)}%` }"
                        ></div>
                        </div>
                        <p class="text-xs text-gray-400 mt-1">
                        Total Load: {{ store.utilizationData.percent.toFixed(1) }}% from Annual Capacity.
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
                        v-model="store.form.notes" 
                        maxlength="500"
                        :rows="5"
                        placeholder="Example: High Priority - external recommendation..."
                        class="w-full"
                        autoresize
                    ></UTextarea>
                    <div class="flex justify-between mt-1">
                        <span class="text-xs text-gray-400">Use it for Special Audit / Investigation.</span>
                        <span class="text-xs text-gray-400">{{ store.form.notes!.length }}/500</span>
                    </div>
                    </UFormField>
                </div>
                </div>
            
                <div class="px-6 py-4 bg-secondary-50 dark:bg-secondary-900 border-t border-secondary-200 dark:border-secondary-700 rounded-b-xl flex justify-end gap-3">
                <UButton 
                    :label="store.isEditing ? 'Update Plan' : 'Save Plan'" 
                    color="primary" 
                    :disabled="!!store.quarterAlert || store.utilizationData.color === 'red'"
                    @click="store.handleSubmit" 
                />
                </div>
            
            </div>
            </UForm>  
            </div>
        </div>
        </div>
    </Teleport>
</template>    

<script setup lang="ts">
import { useAnnualPlanStore } from '~/stores/annual-audit'

// Cukup inisialisasi store. Komponen akan otomatis membaca status showModal, data form, dan fungsi dari sini.
const store = useAnnualPlanStore()
</script>