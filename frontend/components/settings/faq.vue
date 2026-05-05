<template>
    <UCard :ui="{ header: 'flex flex-col gap-4', body: 'p-0' }">
      <template #header>
        <div class="justify-start text-left">
            <h3 class="text-lg font-semibold text-gray-900">Frequently Asked Questions</h3>
            <h6 class="text-sm tracking-wider">Temukan jawaban atas pertanyaan yang sering diajukan</h6>
        </div>
        <div class="justify-start text-left">
            <UInput v-model="searchQuery" icon="i-lucide-search" size="xl" variant="outline" placeholder="Cari topik..." :ui="{
                root: ''
            }"></UInput>

            <div class="flex flex-wrap gap-2 mt-3">
                <p class="text-sm font-medium text-neutral-700">Kategori: </p>
                <UBadge
                    v-for="category in faqCategories"
                    :key="category"
                    :color="selectedCategory === category ? 'primary' : 'neutral'"
                    :variant="selectedCategory === category ? 'solid' : 'subtle'"
                    class="cursor-pointer hover:opacity-80 transition-opacity"
                    @click="selectedCategory = category"
                >
                    {{ category }}
                </UBadge>
            </div>
        </div>
      </template>

      <div class="p-4">
        <UAccordion :items="filteredFaqs" multiple variant="subtle" color="primary">
          <template #default="{ item, open }">
            <UButton color="neutral" variant="ghost" class="w-full justify-between">                                                                                                                      
              <span class="font-medium ">{{ item.label }}</span>                                                                                                                      
              <UIcon :name="open ? 'i-lucide-chevron-up' : 'i-lucide-chevron-down'" class="size-5" />
            </UButton>
          </template>
          <template #content="{ item }">
            <p class="text-sm text-gray-600 mt-2">{{ item.content }}</p>
            
          </template>
        </UAccordion>
      </div>
    </UCard>
</template>
<script setup>
const searchQuery = ref('')
const selectedCategory = ref('All')

const filteredFaqs = computed(() => {
    let result = faqs.value

    // Filter by search query (label)
    if (searchQuery.value) {
        const query = searchQuery.value.toLowerCase()
        result = result.filter(faq =>
            faq.label.toLowerCase().includes(query)
        )
    }

    // Filter by category
    if (selectedCategory.value !== 'All') {
        result = result.filter(faq => {
            return faq.category === selectedCategory.value
        })
    }

    return result
})

const faqCategories = [
    'All',
    'General',
    'Account',
    'Security',
    'Audit',
    'Risk'
]
const faqs = ref(
  [
    {
      label: "How do I change my password?",
      content: "Go to Settings > Security and click on 'Change Password'. You'll need to enter your current password and then create a new one.",
      category: "Security"
    },
    {
      label: "How do I create a new risk profile?",
      content: "Navigate to Risk Management > Risk Profiles and click on the 'Add New Profile' button. Fill in the required fields including risk name, category, impact level, and probability.",
      category: "Risk"
    },
    {
      label: "Can I export audit reports to PDF?",
      content: "Yes, you can export any audit report to PDF. Open the report you want to export and click on the 'Export' button in the top right corner, then select PDF format.",
      category: "Audit"
    },
    {
      label: "How do I change the language?",
      content: "You can change the language from the Settings page or by using the language selector in the top navigation menu. Currently, we support English and Indonesian.",
      category: "General"
    },
    {
      label: "What are the different risk levels?",
      content: "Risk levels are calculated based on impact and probability scores: Low (1-3), Low to Moderate (4-6), Moderate (7-9), Moderate to High (10-12), and High (13-15).",
      category: "Risk"
    },
  ]
)

</script>