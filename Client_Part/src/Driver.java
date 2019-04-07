import java.io.BufferedReader;
import java.io.FileReader;
import java.io.IOException;
import java.util.LinkedList;

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
		LinkedList<Integer> nonceList = new LinkedList<Integer>();
		
		
		try {
			Object obj = new JSONParser().parse(new FileReader(cwd.concat(resume_path)));
			JSONObject resumeJSON = (JSONObject) obj;
			int nonce = 0;
			while(true) {
				int hash = resumeJSON.toString().concat(Integer.toString(nonce)).hashCode();
				if(Integer.toString(hash).startsWith("00000")) {
					nonceList.add(nonce);
				}
				
				nonce++;
			}
			
			
			
			
			
			
			
		} catch (IOException | ParseException e) {
			System.err.println("error during reading resume.json, please check your file and try again");
			e.printStackTrace();
		}
			
		
		
		
		
		
		
		
		
	}

}
