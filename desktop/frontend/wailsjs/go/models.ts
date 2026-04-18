export namespace database {
	
	export class DailyQualityTrend {
	    day: string;
	    ok_qty: number;
	    ng_qty: number;
	    quality_rate: number;
	
	    static createFrom(source: any = {}) {
	        return new DailyQualityTrend(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.day = source["day"];
	        this.ok_qty = source["ok_qty"];
	        this.ng_qty = source["ng_qty"];
	        this.quality_rate = source["quality_rate"];
	    }
	}
	export class DeviceQualityStat {
	    device_id: number;
	    device_name: string;
	    total_qty: number;
	    ok_qty: number;
	    ng_qty: number;
	    quality_rate: number;
	
	    static createFrom(source: any = {}) {
	        return new DeviceQualityStat(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.device_id = source["device_id"];
	        this.device_name = source["device_name"];
	        this.total_qty = source["total_qty"];
	        this.ok_qty = source["ok_qty"];
	        this.ng_qty = source["ng_qty"];
	        this.quality_rate = source["quality_rate"];
	    }
	}
	export class DeviceUtilizationTrend {
	    hour: number;
	    device_id: number;
	    utilization: number;
	
	    static createFrom(source: any = {}) {
	        return new DeviceUtilizationTrend(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.hour = source["hour"];
	        this.device_id = source["device_id"];
	        this.utilization = source["utilization"];
	    }
	}
	export class DeviceVarConfig {
	    device_name: string;
	    production_var_id: number;
	    ng_add_var_id: number;
	    ng_sub_var_id: number;
	
	    static createFrom(source: any = {}) {
	        return new DeviceVarConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.device_name = source["device_name"];
	        this.production_var_id = source["production_var_id"];
	        this.ng_add_var_id = source["ng_add_var_id"];
	        this.ng_sub_var_id = source["ng_sub_var_id"];
	    }
	}
	export class ErrorCode {
	    ErrorCode: number;
	    ErrorMsg: string;
	    CreatedAt: time.Time;
	    UpdatedAt: time.Time;
	
	    static createFrom(source: any = {}) {
	        return new ErrorCode(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ErrorCode = source["ErrorCode"];
	        this.ErrorMsg = source["ErrorMsg"];
	        this.CreatedAt = this.convertValues(source["CreatedAt"], time.Time);
	        this.UpdatedAt = this.convertValues(source["UpdatedAt"], time.Time);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class HourlyOEE {
	    time_period: string;
	    device_name: string;
	    cycle_time: number;
	    total_run_sec: number;
	    total_plan_sec: number;
	    total_products: number;
	    availability_pct: number;
	    performance_pct: number;
	    quality_pct: number;
	    oee_pct: number;
	    hour: number;
	
	    static createFrom(source: any = {}) {
	        return new HourlyOEE(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.time_period = source["time_period"];
	        this.device_name = source["device_name"];
	        this.cycle_time = source["cycle_time"];
	        this.total_run_sec = source["total_run_sec"];
	        this.total_plan_sec = source["total_plan_sec"];
	        this.total_products = source["total_products"];
	        this.availability_pct = source["availability_pct"];
	        this.performance_pct = source["performance_pct"];
	        this.quality_pct = source["quality_pct"];
	        this.oee_pct = source["oee_pct"];
	        this.hour = source["hour"];
	    }
	}
	export class HourlyProduction {
	    hour: number;
	    device_id: number;
	    ok_qty: number;
	    ng_qty: number;
	    total_qty: number;
	
	    static createFrom(source: any = {}) {
	        return new HourlyProduction(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.hour = source["hour"];
	        this.device_id = source["device_id"];
	        this.ok_qty = source["ok_qty"];
	        this.ng_qty = source["ng_qty"];
	        this.total_qty = source["total_qty"];
	    }
	}
	export class HourlyProductionAccurate {
	    time_slot: string;
	    device_name: string;
	    total_qty: number;
	    ng_qty: number;
	    ok_qty: number;
	    quality_rate: number;
	
	    static createFrom(source: any = {}) {
	        return new HourlyProductionAccurate(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.time_slot = source["time_slot"];
	        this.device_name = source["device_name"];
	        this.total_qty = source["total_qty"];
	        this.ng_qty = source["ng_qty"];
	        this.ok_qty = source["ok_qty"];
	        this.quality_rate = source["quality_rate"];
	    }
	}
	export class HourlyProductionPulse {
	    device_id: number;
	    device_name: string;
	    hour: number;
	    ok_qty: number;
	    ng_qty: number;
	    total_qty: number;
	
	    static createFrom(source: any = {}) {
	        return new HourlyProductionPulse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.device_id = source["device_id"];
	        this.device_name = source["device_name"];
	        this.hour = source["hour"];
	        this.ok_qty = source["ok_qty"];
	        this.ng_qty = source["ng_qty"];
	        this.total_qty = source["total_qty"];
	    }
	}
	export class MonthlyProductionAccurate {
	    device_name: string;
	    total_qty: number;
	    ng_qty: number;
	    ok_qty: number;
	    quality_rate: number;
	
	    static createFrom(source: any = {}) {
	        return new MonthlyProductionAccurate(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.device_name = source["device_name"];
	        this.total_qty = source["total_qty"];
	        this.ng_qty = source["ng_qty"];
	        this.ok_qty = source["ok_qty"];
	        this.quality_rate = source["quality_rate"];
	    }
	}
	export class StaffEfficiency {
	    staff_id: number;
	    staff_name: string;
	    total_ok_qty: number;
	    total_ng_qty: number;
	    total_qty: number;
	    quality_rate: number;
	    working_min: number;
	    efficiency: number;
	
	    static createFrom(source: any = {}) {
	        return new StaffEfficiency(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.staff_id = source["staff_id"];
	        this.staff_name = source["staff_name"];
	        this.total_ok_qty = source["total_ok_qty"];
	        this.total_ng_qty = source["total_ng_qty"];
	        this.total_qty = source["total_qty"];
	        this.quality_rate = source["quality_rate"];
	        this.working_min = source["working_min"];
	        this.efficiency = source["efficiency"];
	    }
	}
	export class VariableRow {
	    ID: number;
	    DeviceID: number;
	    VarName: string;
	    DisplayName?: string;
	    DataType?: string;
	    RWMode?: string;
	    Unit?: string;
	    JSONPath: string;
	    ScaleFactor: number;
	    OffsetVal: number;
	    AlarmEnable: boolean;
	    LimitHH?: number;
	    LimitH?: number;
	    LimitL?: number;
	    LimitLL?: number;
	    Deadband?: number;
	    AlarmMsg?: string;
	    StoreMode: number;
	    StoreCycle: number;
	    StoreDeadband: number;
	    SuspiciousValue?: number;
	    DebounceThreshold?: number;
	    StartupSnapshotEnable?: number;
	
	    static createFrom(source: any = {}) {
	        return new VariableRow(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.DeviceID = source["DeviceID"];
	        this.VarName = source["VarName"];
	        this.DisplayName = source["DisplayName"];
	        this.DataType = source["DataType"];
	        this.RWMode = source["RWMode"];
	        this.Unit = source["Unit"];
	        this.JSONPath = source["JSONPath"];
	        this.ScaleFactor = source["ScaleFactor"];
	        this.OffsetVal = source["OffsetVal"];
	        this.AlarmEnable = source["AlarmEnable"];
	        this.LimitHH = source["LimitHH"];
	        this.LimitH = source["LimitH"];
	        this.LimitL = source["LimitL"];
	        this.LimitLL = source["LimitLL"];
	        this.Deadband = source["Deadband"];
	        this.AlarmMsg = source["AlarmMsg"];
	        this.StoreMode = source["StoreMode"];
	        this.StoreCycle = source["StoreCycle"];
	        this.StoreDeadband = source["StoreDeadband"];
	        this.SuspiciousValue = source["SuspiciousValue"];
	        this.DebounceThreshold = source["DebounceThreshold"];
	        this.StartupSnapshotEnable = source["StartupSnapshotEnable"];
	    }
	}

}

export namespace main {
	
	export class RelevantDoc {
	    id: string;
	    content: string;
	    source: string;
	    similarity: number;
	
	    static createFrom(source: any = {}) {
	        return new RelevantDoc(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.content = source["content"];
	        this.source = source["source"];
	        this.similarity = source["similarity"];
	    }
	}
	export class AIQueryResponse {
	    answer: string;
	    relevant_docs: RelevantDoc[];
	    context_used: boolean;
	    error?: string;
	
	    static createFrom(source: any = {}) {
	        return new AIQueryResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.answer = source["answer"];
	        this.relevant_docs = this.convertValues(source["relevant_docs"], RelevantDoc);
	        this.context_used = source["context_used"];
	        this.error = source["error"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class AddKnowledgeResponse {
	    success: boolean;
	    message: string;
	    knowledge_id?: string;
	
	    static createFrom(source: any = {}) {
	        return new AddKnowledgeResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.message = source["message"];
	        this.knowledge_id = source["knowledge_id"];
	    }
	}
	export class AlarmRecordData {
	    id: number;
	    var_id: number;
	    var_name: string;
	    val: number;
	    alarm_type: string;
	    limit_value: number;
	    msg: string;
	    start_time: time.Time;
	    end_time?: time.Time;
	    ack_status: number;
	    duration: string;
	
	    static createFrom(source: any = {}) {
	        return new AlarmRecordData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.var_id = source["var_id"];
	        this.var_name = source["var_name"];
	        this.val = source["val"];
	        this.alarm_type = source["alarm_type"];
	        this.limit_value = source["limit_value"];
	        this.msg = source["msg"];
	        this.start_time = this.convertValues(source["start_time"], time.Time);
	        this.end_time = this.convertValues(source["end_time"], time.Time);
	        this.ack_status = source["ack_status"];
	        this.duration = source["duration"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class BatchRegenerateResult {
	    date: string;
	    shift_id: number;
	    device_id: number;
	    ok: boolean;
	    error: string;
	
	    static createFrom(source: any = {}) {
	        return new BatchRegenerateResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.date = source["date"];
	        this.shift_id = source["shift_id"];
	        this.device_id = source["device_id"];
	        this.ok = source["ok"];
	        this.error = source["error"];
	    }
	}
	export class BreakTime {
	    id: number;
	    name: string;
	    start_hour: number;
	    start_min: number;
	    end_hour: number;
	    end_min: number;
	
	    static createFrom(source: any = {}) {
	        return new BreakTime(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.start_hour = source["start_hour"];
	        this.start_min = source["start_min"];
	        this.end_hour = source["end_hour"];
	        this.end_min = source["end_min"];
	    }
	}
	export class DeleteKnowledgeResponse {
	    success: boolean;
	    message: string;
	
	    static createFrom(source: any = {}) {
	        return new DeleteKnowledgeResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.message = source["message"];
	    }
	}
	export class DeviceEnergyData {
	    device_id: number;
	    device_name: string;
	    real_time_power: number;
	    today_consumption: number;
	    power_unit: string;
	    energy_unit: string;
	
	    static createFrom(source: any = {}) {
	        return new DeviceEnergyData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.device_id = source["device_id"];
	        this.device_name = source["device_name"];
	        this.real_time_power = source["real_time_power"];
	        this.today_consumption = source["today_consumption"];
	        this.power_unit = source["power_unit"];
	        this.energy_unit = source["energy_unit"];
	    }
	}
	export class DeviceStatusData {
	    device_id: number;
	    device_name: string;
	    device_code: string;
	    current_status: number;
	    status_name: string;
	    start_time?: time.Time;
	    duration_min: number;
	    running_min: number;
	    idle_min: number;
	    fault_min: number;
	    utilization: number;
	    operators: string;
	    record_time: string;
	    temperature: string;
	    humidity: string;
	    remark: string;
	
	    static createFrom(source: any = {}) {
	        return new DeviceStatusData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.device_id = source["device_id"];
	        this.device_name = source["device_name"];
	        this.device_code = source["device_code"];
	        this.current_status = source["current_status"];
	        this.status_name = source["status_name"];
	        this.start_time = this.convertValues(source["start_time"], time.Time);
	        this.duration_min = source["duration_min"];
	        this.running_min = source["running_min"];
	        this.idle_min = source["idle_min"];
	        this.fault_min = source["fault_min"];
	        this.utilization = source["utilization"];
	        this.operators = source["operators"];
	        this.record_time = source["record_time"];
	        this.temperature = source["temperature"];
	        this.humidity = source["humidity"];
	        this.remark = source["remark"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class DeviceStatusHistoryData {
	    id: number;
	    device_id: number;
	    status: number;
	    start_time: time.Time;
	    end_time?: time.Time;
	    duration_min: number;
	    extra_data?: string;
	    remark?: string;
	    device?: models.SysDevice;
	    team_name: string;
	    operators: string;
	
	    static createFrom(source: any = {}) {
	        return new DeviceStatusHistoryData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.device_id = source["device_id"];
	        this.status = source["status"];
	        this.start_time = this.convertValues(source["start_time"], time.Time);
	        this.end_time = this.convertValues(source["end_time"], time.Time);
	        this.duration_min = source["duration_min"];
	        this.extra_data = source["extra_data"];
	        this.remark = source["remark"];
	        this.device = this.convertValues(source["device"], models.SysDevice);
	        this.team_name = source["team_name"];
	        this.operators = source["operators"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class Gateway {
	    id: number;
	    gw_name: string;
	    status: number;
	
	    static createFrom(source: any = {}) {
	        return new Gateway(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.gw_name = source["gw_name"];
	        this.status = source["status"];
	    }
	}
	export class HistoryRecord {
	    timestamp: string;
	    value: any;
	
	    static createFrom(source: any = {}) {
	        return new HistoryRecord(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.timestamp = source["timestamp"];
	        this.value = source["value"];
	    }
	}
	export class HistoryDataResponse {
	    records: HistoryRecord[];
	    total: number;
	
	    static createFrom(source: any = {}) {
	        return new HistoryDataResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.records = this.convertValues(source["records"], HistoryRecord);
	        this.total = source["total"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	export class HourlyAlarmCount {
	    hour: number;
	    alarm_count: number;
	    time_slot: string;
	
	    static createFrom(source: any = {}) {
	        return new HourlyAlarmCount(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.hour = source["hour"];
	        this.alarm_count = source["alarm_count"];
	        this.time_slot = source["time_slot"];
	    }
	}
	export class KnowledgeItem {
	    id: string;
	    content: string;
	    source?: string;
	    created_at?: string;
	
	    static createFrom(source: any = {}) {
	        return new KnowledgeItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.content = source["content"];
	        this.source = source["source"];
	        this.created_at = source["created_at"];
	    }
	}
	export class KnowledgeListResponse {
	    success: boolean;
	    message: string;
	    data: KnowledgeItem[];
	    total: number;
	
	    static createFrom(source: any = {}) {
	        return new KnowledgeListResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.message = source["message"];
	        this.data = this.convertValues(source["data"], KnowledgeItem);
	        this.total = source["total"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class ShiftBreak {
	    id: number;
	    shift_id: number;
	    name: string;
	    start_hour: number;
	    start_min: number;
	    end_hour: number;
	    end_min: number;
	
	    static createFrom(source: any = {}) {
	        return new ShiftBreak(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.shift_id = source["shift_id"];
	        this.name = source["name"];
	        this.start_hour = source["start_hour"];
	        this.start_min = source["start_min"];
	        this.end_hour = source["end_hour"];
	        this.end_min = source["end_min"];
	    }
	}
	export class LogicalDayShift {
	    id: number;
	    schedule_id: number;
	    name: string;
	    start_hour: number;
	    start_min: number;
	    end_hour: number;
	    end_min: number;
	    is_active: boolean;
	    sort_order: number;
	    breaks: ShiftBreak[];
	    has_arrived: boolean;
	    is_current: boolean;
	    logical_date: string;
	    calendar_day_offset: number;
	
	    static createFrom(source: any = {}) {
	        return new LogicalDayShift(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.schedule_id = source["schedule_id"];
	        this.name = source["name"];
	        this.start_hour = source["start_hour"];
	        this.start_min = source["start_min"];
	        this.end_hour = source["end_hour"];
	        this.end_min = source["end_min"];
	        this.is_active = source["is_active"];
	        this.sort_order = source["sort_order"];
	        this.breaks = this.convertValues(source["breaks"], ShiftBreak);
	        this.has_arrived = source["has_arrived"];
	        this.is_current = source["is_current"];
	        this.logical_date = source["logical_date"];
	        this.calendar_day_offset = source["calendar_day_offset"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	
	export class ShiftConfig {
	    id: number;
	    schedule_id: number;
	    name: string;
	    start_hour: number;
	    start_min: number;
	    end_hour: number;
	    end_min: number;
	    is_active: boolean;
	    sort_order: number;
	    breaks: ShiftBreak[];
	
	    static createFrom(source: any = {}) {
	        return new ShiftConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.schedule_id = source["schedule_id"];
	        this.name = source["name"];
	        this.start_hour = source["start_hour"];
	        this.start_min = source["start_min"];
	        this.end_hour = source["end_hour"];
	        this.end_min = source["end_min"];
	        this.is_active = source["is_active"];
	        this.sort_order = source["sort_order"];
	        this.breaks = this.convertValues(source["breaks"], ShiftBreak);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class ShiftDeviceOEE {
	    device_name: string;
	    availability_pct: number;
	    performance_pct: number;
	    quality_pct: number;
	    oee_pct: number;
	
	    static createFrom(source: any = {}) {
	        return new ShiftDeviceOEE(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.device_name = source["device_name"];
	        this.availability_pct = source["availability_pct"];
	        this.performance_pct = source["performance_pct"];
	        this.quality_pct = source["quality_pct"];
	        this.oee_pct = source["oee_pct"];
	    }
	}
	export class ShiftOEESummary {
	    shift_name: string;
	    start_label: string;
	    end_label: string;
	    is_current: boolean;
	    has_arrived: boolean;
	    devices: ShiftDeviceOEE[];
	
	    static createFrom(source: any = {}) {
	        return new ShiftOEESummary(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.shift_name = source["shift_name"];
	        this.start_label = source["start_label"];
	        this.end_label = source["end_label"];
	        this.is_current = source["is_current"];
	        this.has_arrived = source["has_arrived"];
	        this.devices = this.convertValues(source["devices"], ShiftDeviceOEE);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class ShiftQualitySummary {
	    shift_name: string;
	    start_label: string;
	    end_label: string;
	    is_current: boolean;
	    has_arrived: boolean;
	    devices: database.DeviceQualityStat[];
	
	    static createFrom(source: any = {}) {
	        return new ShiftQualitySummary(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.shift_name = source["shift_name"];
	        this.start_label = source["start_label"];
	        this.end_label = source["end_label"];
	        this.is_current = source["is_current"];
	        this.has_arrived = source["has_arrived"];
	        this.devices = this.convertValues(source["devices"], database.DeviceQualityStat);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class ShiftScheduleConfig {
	    id: number;
	    name: string;
	    sort_order: number;
	    is_active: boolean;
	    shifts: ShiftConfig[];
	
	    static createFrom(source: any = {}) {
	        return new ShiftScheduleConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.sort_order = source["sort_order"];
	        this.is_active = source["is_active"];
	        this.shifts = this.convertValues(source["shifts"], ShiftConfig);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class TagData {
	    var_name: string;
	    display_name: string;
	    data_type: string;
	    value: string;
	    unit: string;
	    alarm_state: string;
	
	    static createFrom(source: any = {}) {
	        return new TagData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.var_name = source["var_name"];
	        this.display_name = source["display_name"];
	        this.data_type = source["data_type"];
	        this.value = source["value"];
	        this.unit = source["unit"];
	        this.alarm_state = source["alarm_state"];
	    }
	}
	export class TagInfo {
	    var_id: number;
	    var_name: string;
	    display_name: string;
	    unit: string;
	    store_mode: number;
	    data_type: string;
	
	    static createFrom(source: any = {}) {
	        return new TagInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.var_id = source["var_id"];
	        this.var_name = source["var_name"];
	        this.display_name = source["display_name"];
	        this.unit = source["unit"];
	        this.store_mode = source["store_mode"];
	        this.data_type = source["data_type"];
	    }
	}
	export class UserConfig {
	    production_coefficient: number;
	    daily_work_minutes: number;
	    break_times: BreakTime[];
	
	    static createFrom(source: any = {}) {
	        return new UserConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.production_coefficient = source["production_coefficient"];
	        this.daily_work_minutes = source["daily_work_minutes"];
	        this.break_times = this.convertValues(source["break_times"], BreakTime);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class VariableOption {
	    var_id: number;
	    var_name: string;
	    display_name: string;
	    device_name: string;
	    gateway_name: string;
	
	    static createFrom(source: any = {}) {
	        return new VariableOption(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.var_id = source["var_id"];
	        this.var_name = source["var_name"];
	        this.display_name = source["display_name"];
	        this.device_name = source["device_name"];
	        this.gateway_name = source["gateway_name"];
	    }
	}

}

export namespace models {
	
	export class SysTeam {
	    id: number;
	    team_name: string;
	    leader_name?: string;
	    status: number;
	    created_at: time.Time;
	
	    static createFrom(source: any = {}) {
	        return new SysTeam(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.team_name = source["team_name"];
	        this.leader_name = source["leader_name"];
	        this.status = source["status"];
	        this.created_at = this.convertValues(source["created_at"], time.Time);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class ProMachineSession {
	    id: number;
	    device_id: number;
	    team_id: number;
	    staff_ids: string;
	    login_time: time.Time;
	    logout_time?: time.Time;
	    duration_min: number;
	    team?: SysTeam;
	
	    static createFrom(source: any = {}) {
	        return new ProMachineSession(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.device_id = source["device_id"];
	        this.team_id = source["team_id"];
	        this.staff_ids = source["staff_ids"];
	        this.login_time = this.convertValues(source["login_time"], time.Time);
	        this.logout_time = this.convertValues(source["logout_time"], time.Time);
	        this.duration_min = source["duration_min"];
	        this.team = this.convertValues(source["team"], SysTeam);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class ProOrder {
	    id: number;
	    order_no: string;
	    product_code: string;
	    target_device_id?: number;
	    plan_qty: number;
	    actual_qty: number;
	    ok_qty: number;
	    ng_qty: number;
	    status: number;
	    start_time?: time.Time;
	    end_time?: time.Time;
	    used_seconds: number;
	    current_start_time?: time.Time;
	    version: number;
	    created_at: time.Time;
	
	    static createFrom(source: any = {}) {
	        return new ProOrder(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.order_no = source["order_no"];
	        this.product_code = source["product_code"];
	        this.target_device_id = source["target_device_id"];
	        this.plan_qty = source["plan_qty"];
	        this.actual_qty = source["actual_qty"];
	        this.ok_qty = source["ok_qty"];
	        this.ng_qty = source["ng_qty"];
	        this.status = source["status"];
	        this.start_time = this.convertValues(source["start_time"], time.Time);
	        this.end_time = this.convertValues(source["end_time"], time.Time);
	        this.used_seconds = source["used_seconds"];
	        this.current_start_time = this.convertValues(source["current_start_time"], time.Time);
	        this.version = source["version"];
	        this.created_at = this.convertValues(source["created_at"], time.Time);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class ProProductionRun {
	    id: number;
	    order_id: number;
	    device_id: number;
	    team_id: number;
	    run_ok_qty: number;
	    run_ng_qty: number;
	    start_time: time.Time;
	    end_time?: time.Time;
	    operator_ids: string;
	    remark?: string;
	    order?: ProOrder;
	    team?: SysTeam;
	
	    static createFrom(source: any = {}) {
	        return new ProProductionRun(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.order_id = source["order_id"];
	        this.device_id = source["device_id"];
	        this.team_id = source["team_id"];
	        this.run_ok_qty = source["run_ok_qty"];
	        this.run_ng_qty = source["run_ng_qty"];
	        this.start_time = this.convertValues(source["start_time"], time.Time);
	        this.end_time = this.convertValues(source["end_time"], time.Time);
	        this.operator_ids = source["operator_ids"];
	        this.remark = source["remark"];
	        this.order = this.convertValues(source["order"], ProOrder);
	        this.team = this.convertValues(source["team"], SysTeam);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class ProShiftSnapshot {
	    id: number;
	    snapshot_date: string;
	    device_id: number;
	    device_name: string;
	    schedule_id: number;
	    shift_id: number;
	    shift_name: string;
	    shift_start: time.Time;
	    shift_end: time.Time;
	    break_config: string;
	    cycle_time: number;
	    plan_work_sec: number;
	    total_qty: number;
	    ok_qty: number;
	    ng_qty: number;
	    device_run_sec: number;
	    device_idle_sec: number;
	    device_fault_sec: number;
	    team_id?: number;
	    team_name: string;
	    staff_snapshot: string;
	    availability_pct: number;
	    performance_pct: number;
	    quality_pct: number;
	    oee_pct: number;
	    session_id?: number;
	    sessions_snapshot: string;
	    created_at: time.Time;
	
	    static createFrom(source: any = {}) {
	        return new ProShiftSnapshot(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.snapshot_date = source["snapshot_date"];
	        this.device_id = source["device_id"];
	        this.device_name = source["device_name"];
	        this.schedule_id = source["schedule_id"];
	        this.shift_id = source["shift_id"];
	        this.shift_name = source["shift_name"];
	        this.shift_start = this.convertValues(source["shift_start"], time.Time);
	        this.shift_end = this.convertValues(source["shift_end"], time.Time);
	        this.break_config = source["break_config"];
	        this.cycle_time = source["cycle_time"];
	        this.plan_work_sec = source["plan_work_sec"];
	        this.total_qty = source["total_qty"];
	        this.ok_qty = source["ok_qty"];
	        this.ng_qty = source["ng_qty"];
	        this.device_run_sec = source["device_run_sec"];
	        this.device_idle_sec = source["device_idle_sec"];
	        this.device_fault_sec = source["device_fault_sec"];
	        this.team_id = source["team_id"];
	        this.team_name = source["team_name"];
	        this.staff_snapshot = source["staff_snapshot"];
	        this.availability_pct = source["availability_pct"];
	        this.performance_pct = source["performance_pct"];
	        this.quality_pct = source["quality_pct"];
	        this.oee_pct = source["oee_pct"];
	        this.session_id = source["session_id"];
	        this.sessions_snapshot = source["sessions_snapshot"];
	        this.created_at = this.convertValues(source["created_at"], time.Time);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class SysStaff {
	    id: number;
	    staff_code: string;
	    name: string;
	    current_team_id?: number;
	    is_active: number;
	    created_at: time.Time;
	    current_team?: SysTeam;
	
	    static createFrom(source: any = {}) {
	        return new SysStaff(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.staff_code = source["staff_code"];
	        this.name = source["name"];
	        this.current_team_id = source["current_team_id"];
	        this.is_active = source["is_active"];
	        this.created_at = this.convertValues(source["created_at"], time.Time);
	        this.current_team = this.convertValues(source["current_team"], SysTeam);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class SessionStatusResponse {
	    id: number;
	    device_id: number;
	    team_id: number;
	    staff_ids: string;
	    login_time: time.Time;
	    logout_time?: time.Time;
	    duration_min: number;
	    team?: SysTeam;
	    is_active: boolean;
	    staff_list: SysStaff[];
	    worked_min: number;
	    idle_min: number;
	    efficiency: number;
	
	    static createFrom(source: any = {}) {
	        return new SessionStatusResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.device_id = source["device_id"];
	        this.team_id = source["team_id"];
	        this.staff_ids = source["staff_ids"];
	        this.login_time = this.convertValues(source["login_time"], time.Time);
	        this.logout_time = this.convertValues(source["logout_time"], time.Time);
	        this.duration_min = source["duration_min"];
	        this.team = this.convertValues(source["team"], SysTeam);
	        this.is_active = source["is_active"];
	        this.staff_list = this.convertValues(source["staff_list"], SysStaff);
	        this.worked_min = source["worked_min"];
	        this.idle_min = source["idle_min"];
	        this.efficiency = source["efficiency"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class SysDevice {
	    id: number;
	    gateway_id: number;
	    device_code: string;
	    device_name: string;
	    identify_key?: string;
	    schedule_id?: number;
	    cycle_time?: number;
	
	    static createFrom(source: any = {}) {
	        return new SysDevice(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.gateway_id = source["gateway_id"];
	        this.device_code = source["device_code"];
	        this.device_name = source["device_name"];
	        this.identify_key = source["identify_key"];
	        this.schedule_id = source["schedule_id"];
	        this.cycle_time = source["cycle_time"];
	    }
	}
	export class SysDeviceStatus {
	    id: number;
	    device_id: number;
	    status: number;
	    start_time: time.Time;
	    end_time?: time.Time;
	    duration_min: number;
	    extra_data?: string;
	    remark?: string;
	    device?: SysDevice;
	
	    static createFrom(source: any = {}) {
	        return new SysDeviceStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.device_id = source["device_id"];
	        this.status = source["status"];
	        this.start_time = this.convertValues(source["start_time"], time.Time);
	        this.end_time = this.convertValues(source["end_time"], time.Time);
	        this.duration_min = source["duration_min"];
	        this.extra_data = source["extra_data"];
	        this.remark = source["remark"];
	        this.device = this.convertValues(source["device"], SysDevice);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	export class SysStaffHistory {
	    id: number;
	    staff_id: number;
	    team_id: number;
	    action_type: number;
	    happened_at: time.Time;
	    operator_name?: string;
	    staff?: SysStaff;
	    team?: SysTeam;
	
	    static createFrom(source: any = {}) {
	        return new SysStaffHistory(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.staff_id = source["staff_id"];
	        this.team_id = source["team_id"];
	        this.action_type = source["action_type"];
	        this.happened_at = this.convertValues(source["happened_at"], time.Time);
	        this.operator_name = source["operator_name"];
	        this.staff = this.convertValues(source["staff"], SysStaff);
	        this.team = this.convertValues(source["team"], SysTeam);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	export class Task {
	    task_id: number;
	    task_name: string;
	    task_type: number;
	    is_enabled: boolean;
	    description: string;
	    created_at: time.Time;
	    updated_at: time.Time;
	    cron_expr: string;
	    interval_sec: number;
	    last_run_time: time.Time;
	    trigger_var_id: number;
	    trigger_var_name: string;
	    change_type: string;
	    change_threshold: number;
	    condition_expr: string;
	    action_type: number;
	    action_config: string;
	
	    static createFrom(source: any = {}) {
	        return new Task(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.task_id = source["task_id"];
	        this.task_name = source["task_name"];
	        this.task_type = source["task_type"];
	        this.is_enabled = source["is_enabled"];
	        this.description = source["description"];
	        this.created_at = this.convertValues(source["created_at"], time.Time);
	        this.updated_at = this.convertValues(source["updated_at"], time.Time);
	        this.cron_expr = source["cron_expr"];
	        this.interval_sec = source["interval_sec"];
	        this.last_run_time = this.convertValues(source["last_run_time"], time.Time);
	        this.trigger_var_id = source["trigger_var_id"];
	        this.trigger_var_name = source["trigger_var_name"];
	        this.change_type = source["change_type"];
	        this.change_threshold = source["change_threshold"];
	        this.condition_expr = source["condition_expr"];
	        this.action_type = source["action_type"];
	        this.action_config = source["action_config"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

export namespace time {
	
	export class Time {
	
	
	    static createFrom(source: any = {}) {
	        return new Time(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	
	    }
	}

}

