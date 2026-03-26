// 大屏幕模拟数据 - 用于演示
// 使用方法：在 Cockpit.vue 中设置 USE_MOCK_DATA = true

// 获取当前小时（用于生成到当前时间的数据）
const currentHour = new Date().getHours()

// 订单计划产量（目标，还没完成）
const ORDER_PLAN_QTY = {
  device1: 65,   // 一号焊机计划要生产65件
  device2: 50    // 二号焊机计划要生产50件
}

// 实际已生产数量（良品 + 不良品）
// 设计良品率：约98%
// 一号机：实际生产50件 = 49件良品 + 1件不良 → 良品率 49÷50×100 = 98.00%
// 二号机：实际生产37件 = 36件良品 + 1件不良 → 良品率 36÷37×100 = 97.30%
// 总计：实际生产87件 = 85件良品 + 2件不良 → 良品率 85÷87×100 = 97.70%
const ACTUAL_PRODUCTION = {
  device1: {
    ok_qty: 49,   // 良品数
    ng_qty: 1     // 不良品数
  },
  device2: {
    ok_qty: 36,   // 良品数
    ng_qty: 1     // 不良品数
  }
}

// 生成从7点到19点的完整时间范围（用于演示，显示完整一天）
const generateHourRange = () => {
  const hours = []
  const start = 7
  const end = 19  // 固定到19点，显示完整工作时间
  for (let i = start; i <= end; i++) {
    hours.push(i)
  }
  return hours
}

// 生成每小时产量数据（98%良品率，12点午休跳过）
const generateHourlyProduction = () => {
  const hours = generateHourRange()
  const data = []
  const lastHour = currentHour - 1 // 显示到当前小时的前一个小时
  
  // 计算有效工作小时数（排除12点和未到的小时）
  const workingHours = hours.filter(h => h <= lastHour && h !== 12)
  const workingHourCount = workingHours.length
  
  if (workingHourCount === 0) {
    return data
  }
  
  // 一号焊机 - 按实际产量平均分配到每小时
  const device1OkTotal = ACTUAL_PRODUCTION.device1.ok_qty
  const device1NgTotal = ACTUAL_PRODUCTION.device1.ng_qty
  
  const device1HourlyAvg = Math.floor(device1OkTotal / workingHourCount)
  const device1Remainder = device1OkTotal - (device1HourlyAvg * workingHourCount)
  
  let device1OkDistributed = 0
  let device1NgDistributed = 0
  hours.forEach((hour, index) => {
    // 超过当前小时的前一个小时，不生成数据（前端会显示null）
    if (hour > lastHour) {
      return
    }
    
    // 12点午休时间，添加数据但产量为0（用于图表断开）
    if (hour === 12) {
      data.push({
        hour: hour,
        device_id: 1,
        device_name: '一号焊机',
        ok_qty: 0,
        ng_qty: 0
      })
      return
    }
    
    // 基础良品产量 + 余数分配
    let okQty = device1HourlyAvg
    if (device1OkDistributed < device1Remainder) {
      okQty += 1
      device1OkDistributed++
    }
    
    // 不良品分配（分散到几个小时，确保总数正确）
    let ngQty = 0
    if (device1NgDistributed < device1NgTotal) {
      // 在剩余工作小时中平均分配不良品
      const remainingWorkHours = workingHours.filter(h => h >= hour).length
      const remainingNg = device1NgTotal - device1NgDistributed
      if (remainingNg >= remainingWorkHours || Math.random() < 0.3) {
        ngQty = 1
        device1NgDistributed++
      }
    }
    
    data.push({
      hour: hour,
      device_id: 1,
      device_name: '一号焊机',
      ok_qty: okQty,
      ng_qty: ngQty
    })
  })
  
  // 二号焊机 - 按实际产量平均分配到每小时
  const device2OkTotal = ACTUAL_PRODUCTION.device2.ok_qty
  const device2NgTotal = ACTUAL_PRODUCTION.device2.ng_qty
  
  const device2HourlyAvg = Math.floor(device2OkTotal / workingHourCount)
  const device2Remainder = device2OkTotal - (device2HourlyAvg * workingHourCount)
  
  let device2OkDistributed = 0
  let device2NgDistributed = 0
  hours.forEach((hour, index) => {
    // 超过当前小时的前一个小时，不生成数据（前端会显示null）
    if (hour > lastHour) {
      return
    }
    
    // 12点午休时间，添加数据但产量为0（用于图表断开）
    if (hour === 12) {
      data.push({
        hour: hour,
        device_id: 2,
        device_name: '二号焊机',
        ok_qty: 0,
        ng_qty: 0
      })
      return
    }
    
    // 基础良品产量 + 余数分配
    let okQty = device2HourlyAvg
    if (device2OkDistributed < device2Remainder) {
      okQty += 1
      device2OkDistributed++
    }
    
    // 不良品分配（分散到几个小时，确保总数正确）
    let ngQty = 0
    if (device2NgDistributed < device2NgTotal) {
      // 在剩余工作小时中平均分配不良品
      const remainingWorkHours = workingHours.filter(h => h >= hour).length
      const remainingNg = device2NgTotal - device2NgDistributed
      if (remainingNg >= remainingWorkHours || Math.random() < 0.3) {
        ngQty = 1
        device2NgDistributed++
      }
    }
    
    data.push({
      hour: hour,
      device_id: 2,
      device_name: '二号焊机',
      ok_qty: okQty,
      ng_qty: ngQty
    })
  })
  
  return data
}

// 生成稼动率趋势数据（从7点开始，12点午休不显示，平均92.5%）
const generateUtilizationTrend = () => {
  const hours = generateHourRange()
  const data = []
  const lastHour = currentHour - 1 // 显示到当前小时的前一个小时
  
  // 一号焊机 - 稼动率围绕93%波动（88%-96%）
  hours.forEach(hour => {
    // 超过当前小时的前一个小时，不生成数据（前端会显示null）
    if (hour > lastHour) {
      return
    }
    
    // 12点午休时间，添加数据但稼动率为0（用于图表断开）
    if (hour === 12) {
      data.push({
        hour: hour,
        device_id: 1,
        utilization: 0
      })
      return
    }
    
    let utilization
    if (hour >= 7 && hour < 12) {
      // 上午7-11点：90%-96%（刚开始稍低，逐渐提升）
      utilization = 90 + Math.random() * 6
    } else if (hour >= 13) {
      // 下午13点开始：88%-94%（午休后稍有下降）
      utilization = 88 + Math.random() * 6
    } else {
      utilization = 0
    }
    
    data.push({
      hour: hour,
      device_id: 1,
      utilization: parseFloat(utilization.toFixed(1))
    })
  })
  
  // 二号焊机 - 稼动率围绕92%波动（87%-95%）
  hours.forEach(hour => {
    // 超过当前小时的前一个小时，不生成数据（前端会显示null）
    if (hour > lastHour) {
      return
    }
    
    // 12点午休时间，添加数据但稼动率为0（用于图表断开）
    if (hour === 12) {
      data.push({
        hour: hour,
        device_id: 2,
        utilization: 0
      })
      return
    }
    
    let utilization
    if (hour >= 7 && hour < 12) {
      // 上午7-11点：89%-95%
      utilization = 89 + Math.random() * 6
    } else if (hour >= 13) {
      // 下午13点开始：87%-93%
      utilization = 87 + Math.random() * 6
    } else {
      utilization = 0
    }
    
    data.push({
      hour: hour,
      device_id: 2,
      utilization: parseFloat(utilization.toFixed(1))
    })
  })
  
  return data
}

// 生成模拟数据
const hourlyProductionData = generateHourlyProduction()
const utilizationTrendData = generateUtilizationTrend()

// 从 hourlyProduction 计算实际产量（确保数据一致性）
const device1ActualData = hourlyProductionData
  .filter(h => h.device_id === 1)
  .reduce((acc, h) => ({
    ok_qty: acc.ok_qty + h.ok_qty,
    ng_qty: acc.ng_qty + h.ng_qty
  }), { ok_qty: 0, ng_qty: 0 })

const device2ActualData = hourlyProductionData
  .filter(h => h.device_id === 2)
  .reduce((acc, h) => ({
    ok_qty: acc.ok_qty + h.ok_qty,
    ng_qty: acc.ng_qty + h.ng_qty
  }), { ok_qty: 0, ng_qty: 0 })

export const mockData = {
  // 工单数据 - 从 hourlyProduction 计算得出，确保良品率一致
  orders: [
    {
      id: 1,
      order_no: 'WO20241222001',
      product_code: 'LW-2024-A01',
      target_device_id: 1,
      plan_qty: ORDER_PLAN_QTY.device1,  // 计划数量（目标）
      actual_qty: device1ActualData.ok_qty + device1ActualData.ng_qty,  // 实际产量（良品+不良）
      ok_qty: device1ActualData.ok_qty,  // 良品数
      ng_qty: device1ActualData.ng_qty,  // 不良品数
      status: 1 // 进行中
    },
    {
      id: 2,
      order_no: 'WO20241222002',
      product_code: 'LW-2024-B02',
      target_device_id: 2,
      plan_qty: ORDER_PLAN_QTY.device2,  // 计划数量（目标）
      actual_qty: device2ActualData.ok_qty + device2ActualData.ng_qty,  // 实际产量（良品+不良）
      ok_qty: device2ActualData.ok_qty,  // 良品数
      ng_qty: device2ActualData.ng_qty,  // 不良品数
      status: 1 // 进行中
    }
  ],

  // 设备数据
  devices: [
    {
      device_id: 1,
      device_name: '一号焊机',
      device_code: 'WM-001',
      current_status: 1 // 运行中
    },
    {
      device_id: 2,
      device_name: '二号焊机',
      device_code: 'WM-002',
      current_status: 1 // 运行中
    }
  ],

  // 设备统计
  deviceStats: {
    avg_utilization: 92.5 // 平均稼动率
  },

  // 每小时产量数据（动态生成到当前小时）
  hourlyProduction: hourlyProductionData,

  // 员工稼动率数据 - 使用真实员工姓名（平均92.5%左右）
  staffEfficiency: [
    { staff_id: 10, staff_name: '仲小平', efficiency: 95, working_min: currentHour >= 7 ? (currentHour - 7) * 60 * 0.95 : 0 },
    { staff_id: 11, staff_name: '王家旭', efficiency: 93, working_min: currentHour >= 7 ? (currentHour - 7) * 60 * 0.93 : 0 },
    { staff_id: 12, staff_name: '陈英华', efficiency: 92, working_min: currentHour >= 7 ? (currentHour - 7) * 60 * 0.92 : 0 },
    { staff_id: 13, staff_name: '杨正飞', efficiency: 91, working_min: currentHour >= 7 ? (currentHour - 7) * 60 * 0.91 : 0 },
    { staff_id: 14, staff_name: '毛福东', efficiency: 91, working_min: currentHour >= 7 ? (currentHour - 7) * 60 * 0.91 : 0 }
  ],

  // 稼动率趋势数据（动态生成到当前小时）
  utilizationTrend: utilizationTrendData,

  // 报警数（少量报警，更真实）
  alarmCount: 2,

  // 员工列表
  staffList: [
    { id: 10, staff_code: '89898', name: '仲小平', current_team_id: 3, is_active: 1 },
    { id: 11, staff_code: '10001', name: '王家旭', current_team_id: 3, is_active: 1 },
    { id: 12, staff_code: '10002', name: '陈英华', current_team_id: 3, is_active: 1 },
    { id: 13, staff_code: '10003', name: '杨正飞', current_team_id: 4, is_active: 1 },
    { id: 14, staff_code: '10004', name: '毛福东', current_team_id: 4, is_active: 1 }
  ],

  // 实时数据（用于报警显示，有2个报警）
  realtimeTags: [
    { id: 1, display_name: '一号机温度', value: 85.5, alarm_state: false },
    { id: 2, display_name: '二号机温度', value: 92.8, alarm_state: true }, // 温度偏高报警
    { id: 3, display_name: '一号机压力', value: 6.8, alarm_state: false },
    { id: 4, display_name: '二号机压力', value: 7.2, alarm_state: false },
    { id: 5, display_name: '一号机电流', value: 135.8, alarm_state: true }, // 电流偏高报警
    { id: 6, display_name: '二号机电流', value: 118.5, alarm_state: false }
  ]
}

// 计算统计数据（用于验证）
const totalActualQty = mockData.orders.reduce((sum, order) => sum + order.actual_qty, 0)
const totalOkQty = mockData.orders.reduce((sum, order) => sum + order.ok_qty, 0)
const totalNgQty = mockData.orders.reduce((sum, order) => sum + order.ng_qty, 0)
const qualityRate = totalActualQty > 0 ? ((totalOkQty / totalActualQty) * 100).toFixed(2) : 0

// 验证 hourlyProduction 和 orders 数据一致性
const hourlyTotalOk = hourlyProductionData.reduce((sum, h) => sum + h.ok_qty, 0)
const hourlyTotalNg = hourlyProductionData.reduce((sum, h) => sum + h.ng_qty, 0)
const hourlyTotal = hourlyTotalOk + hourlyTotalNg

// 计算每个工单的良品率（用于验证）
const order1QualityRate = mockData.orders[0].actual_qty > 0 
  ? ((mockData.orders[0].ok_qty / mockData.orders[0].actual_qty) * 100).toFixed(2)
  : 0
const order2QualityRate = mockData.orders[1].actual_qty > 0 
  ? ((mockData.orders[1].ok_qty / mockData.orders[1].actual_qty) * 100).toFixed(2)
  : 0

// 计算完成率
const totalPlanQty = ORDER_PLAN_QTY.device1 + ORDER_PLAN_QTY.device2
const completionRate = totalPlanQty > 0 ? ((totalActualQty / totalPlanQty) * 100).toFixed(2) : 0

// 打印统计信息（方便调试和验证）
console.log('📊 模拟数据统计:')
console.log(`  - 时间范围: 7:00 - 19:00`)
console.log(`  `)
console.log(`  【计划 vs 实际】`)
console.log(`  - 计划产量: 一号机=${ORDER_PLAN_QTY.device1}件, 二号机=${ORDER_PLAN_QTY.device2}件, 合计=${totalPlanQty}件`)
console.log(`  - 实际产量: 一号机=${mockData.orders[0].actual_qty}件, 二号机=${mockData.orders[1].actual_qty}件, 合计=${totalActualQty}件`)
console.log(`  - 完成率: ${completionRate}% (${totalActualQty}÷${totalPlanQty}×100)`)
console.log(`  `)
console.log(`  【良品率计算】`)
console.log(`  - 良品数: 一号机=${mockData.orders[0].ok_qty}件, 二号机=${mockData.orders[1].ok_qty}件, 合计=${totalOkQty}件`)
console.log(`  - 不良品: 一号机=${mockData.orders[0].ng_qty}件, 二号机=${mockData.orders[1].ng_qty}件, 合计=${totalNgQty}件`)
console.log(`  - 良品率: ${qualityRate}% (一号机=${order1QualityRate}%, 二号机=${order2QualityRate}%)`)
console.log(`  - 验证: ${totalOkQty}÷${totalActualQty}×100 = ${qualityRate}%`)
console.log(`  `)
console.log(`  【其他指标】`)
console.log(`  - 平均稼动率: ${mockData.deviceStats.avg_utilization}%`)
console.log(`  - 报警数: ${mockData.alarmCount} 个`)
console.log(`  - 工单数: ${mockData.orders.length} 个`)
console.log(`  - 员工数: ${mockData.staffList.length} 人`)
console.log(`  `)
console.log(`✅ 数据一致性验证: hourlyProduction总计=${hourlyTotal}件(良品${hourlyTotalOk}+不良${hourlyTotalNg}), orders总计=${totalActualQty}件(良品${totalOkQty}+不良${totalNgQty})`)
