
import java.io.FileReader;
import java.io.IOException;
import java.nio.charset.StandardCharsets;
import java.security.MessageDigest;
import java.security.NoSuchAlgorithmException;
import java.util.LinkedList;

import org.json.simple.JSONObject; 
import org.json.simple.parser.*; 


public class Driver {
	
	public static String resume_path = "/resume.json";
	public static int numOfZero = 3;
	
	
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
				MessageDigest digest = MessageDigest.getInstance("SHA-256");
				byte[] hash = digest.digest(resumeJSON.toString().concat(Integer.toString(nonce)).getBytes(StandardCharsets.UTF_8));
				boolean flag = true;
				for(int i = 0; i < numOfZero; i++) {
					if(hash[i] != 0) {
						flag = false;
					}
				}
				if(flag) {
					nonceList.add(nonce);
				}
				nonce++;
			}
			
			
			
			
			
			
			
		} catch (IOException e) {
			System.err.println("file does not exist");
		} catch (NoSuchAlgorithmException e) {
			System.err.println("encryption error, this computer does not have such encryption method");
		} catch (ParseException e) {
			System.err.println("error during paring json file, please check the format and content");
		}
			
		
		
		
		
		
		
		
		
	}

}
