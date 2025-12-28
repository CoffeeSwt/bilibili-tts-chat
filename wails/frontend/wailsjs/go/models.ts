export namespace config {
	
	export class UserConfig {
	    room_id_code: string;
	    room_description: string;
	    assistant_name: string;
	    max_user_data_len: number;
	    cleanup_interval: number;
	    volume: number;
	    speech_rate: number;
	    assistant_memory_size: number;
	    use_llm_replay: boolean;
	    first_start: boolean;
	
	    static createFrom(source: any = {}) {
	        return new UserConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.room_id_code = source["room_id_code"];
	        this.room_description = source["room_description"];
	        this.assistant_name = source["assistant_name"];
	        this.max_user_data_len = source["max_user_data_len"];
	        this.cleanup_interval = source["cleanup_interval"];
	        this.volume = source["volume"];
	        this.speech_rate = source["speech_rate"];
	        this.assistant_memory_size = source["assistant_memory_size"];
	        this.use_llm_replay = source["use_llm_replay"];
	        this.first_start = source["first_start"];
	    }
	}

}

