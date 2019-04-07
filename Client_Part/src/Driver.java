import java.io.BufferedReader;
import java.io.FileReader;
import java.io.IOException;

import org.json.simple.JSONArray; 
import org.json.simple.JSONObject; 
import org.json.simple.parser.*; 


public class Driver {
	
	public static String resume_path = "/resume.json";
	
	
	public static void main(String[] args) {
		// TODO Auto-generated method stub
		
		readResume();
		
		
		
		
		
		
		
		
	}

	/**
	 * Read the resume json file
	 */
	private static void readResume() {
		String cwd = System.getProperty("user.dir");
		
		try {
			Object obj = new JSONParser().parse(new FileReader(cwd.concat(resume_path)));
			JSONObject resumeJSON = (JSONObject) obj;
			while(true) {
				HashCode hash = resumeJSON
			}
			
			
			
			
			
			
			
		} catch (IOException | ParseException e) {
			System.err.println("error during reading resume.json, please check your file and try again");
			e.printStackTrace();
		}
			
		
		
		
		
		
		
		
		
	}

}
