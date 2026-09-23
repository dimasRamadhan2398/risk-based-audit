const monthNames = ["Januari", "Februari", "Maret", "April", "Mei", "Juni", "Juli", "Agustus", "September", "Oktober", "November", "Desember"]

function generateMonitoringChecks(startDateStr, endDateStr) {
    const checks = []
    if (!startDateStr || !endDateStr) return checks

    const start = new Date(startDateStr)
    const end = new Date(endDateStr)
    if (isNaN(start.getTime()) || isNaN(end.getTime())) return checks

    let current = new Date(start)
    let mIndex = 1

    while (current <= end) {
        let mStart = new Date(current.getFullYear(), current.getMonth(), 1)
        if (mStart < start) mStart = new Date(start)

        let mEnd = new Date(current.getFullYear(), current.getMonth() + 1, 0, 23, 59, 59)
        if (mEnd > end) mEnd = new Date(end)

        const weeks = []
        let wCurrent = new Date(mStart)
        let wIndex = 1

        while (wCurrent <= mEnd) {
            let wEnd = new Date(wCurrent)
            wEnd.setDate(wEnd.getDate() + 6)
            wEnd.setHours(23, 59, 59, 999)
            if (wEnd > mEnd) wEnd = new Date(mEnd)

            const wStartFmt = wCurrent.toLocaleDateString('id-ID', { day: '2-digit', month: 'short' })
            const wEndFmt = wEnd.toLocaleDateString('id-ID', { day: '2-digit', month: 'short' })

            weeks.push({
                id: `M${mIndex}W${wIndex}`,
                label: `Minggu ${wIndex} (${wStartFmt} - ${wEndFmt})`,
                checked: false,
                notes: "",
                startDate: wCurrent.toISOString(),
                endDate: wEnd.toISOString()
            })

            wCurrent = new Date(wEnd)
            wCurrent.setTime(wCurrent.getTime() + 1000)
            wIndex++
        }

        checks.push({
            id: `M${mIndex}`,
            label: `${monthNames[current.getMonth()]} ${current.getFullYear()}`,
            checked: false,
            notes: "",
            startDate: mStart.toISOString(),
            endDate: mEnd.toISOString(),
            weeks: weeks
        })

        current = new Date(current.getFullYear(), current.getMonth() + 1, 1)
        mIndex++
    }

    return checks
}

console.log(JSON.stringify(generateMonitoringChecks("2026-09-01T00:00:00.000Z", "2026-09-30T23:59:59.000Z"), null, 2))
